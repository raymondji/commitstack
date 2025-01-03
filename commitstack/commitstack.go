package commitstack

import (
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"sort"
	"strings"

	"github.com/raymondji/git-stack-cli/commitstack/commitgraph"
	"github.com/raymondji/git-stack-cli/libgit"
)

type Commit struct {
	Hash          string
	Author        string
	Subject       string
	Date          string
	LocalBranches []string
}

type DivergenceError struct {
	StackName       string
	OtherStackNames []string
}

func (e DivergenceError) Error() string {
	assert(len(e.OtherStackNames) > 0, "other stack names must not be empty")

	if len(e.OtherStackNames) == 1 {
		return fmt.Sprintf("Stack %s has diverged from stack %s", e.StackName, e.OtherStackNames[0])
	} else {
		return fmt.Sprintf("Stack %s has diverged from stacks %s", e.StackName, strings.Join(e.OtherStackNames, ", "))
	}
}

type BranchCollisionError struct {
	StackName string
	Branches  []string
}

func (e BranchCollisionError) Error() string {
	return fmt.Sprintf(
		"Stack %s contains multiple branches (%v) pointing to the same commit",
		e.StackName, strings.Join(e.Branches, ", "))
}

type MergeCommitError struct {
	MergeCommitHash    string
	ContainingBranches []string
}

func (e MergeCommitError) Error() string {
	switch len(e.ContainingBranches) {
	case 0:
		return fmt.Sprintf("Merge commit %v", e.MergeCommitHash)
	case 1:
		return fmt.Sprintf("Merge commit %v is present in branch %v", e.MergeCommitHash, e.ContainingBranches[0])
	default:
		return fmt.Sprintf("Merge commit %v is present in branches %v", e.MergeCommitHash, strings.Join(e.ContainingBranches, ", "))
	}
}

type Git interface {
	GetBranchesContainingCommit(commitHash string) ([]string, error)
}

type Stack struct {
	Commits []Commit // Ordered from top to bottom, guaranteed to not be empty

	// Validation errors indicate that we were able to infer a commit stack, but the stack isn't valid
	ValidationErrors []error
}

// AllBranches returns all branches associated with all commits in the stack.
// Guaranteed to return at least one branch.
func (s Stack) AllBranches() []string {
	var out []string
	for _, c := range s.Commits {
		if len(c.LocalBranches) == 0 {
			continue
		}
		out = append(out, c.LocalBranches...)
	}
	assert(len(out) > 0, "commitstack must contain at least one branch")
	return out
}

// UniqueBranches returns a list of branch names where
// each branch name must be associated with a unique commit.
// If any commits are associated with multiple branches, this returns an error.
// Guaranteed to return at least one branch.
func (s Stack) UniqueBranches() ([]string, error) {
	var out []string
	for _, c := range s.Commits {
		switch len(c.LocalBranches) {
		case 0:
			continue
		case 1:
			out = append(out, c.LocalBranches[0])
		default:
			assert(len(s.ValidationErrors) > 0, "stack should have a branch collision validation error")
			return nil, fmt.Errorf("stack contains multiple branches")
		}
	}
	return out, nil
}

// Name returns a valid git reference (e.g. a branch ref or commit hash)
func (s Stack) Name() string {
	tip := s.Commits[0]
	if len(tip.LocalBranches) > 0 {
		return tip.LocalBranches[0]
	}
	return tip.Hash
}

func (s Stack) IsCurrent(currentBranch string) bool {
	for _, c := range s.Commits {
		if slices.Contains(c.LocalBranches, currentBranch) {
			return true
		}
	}
	return false
}

func GetCurrent(stacks []Stack, currentBranch string) (Stack, error) {
	var currStacks []Stack
	for _, s := range stacks {
		if s.IsCurrent(currentBranch) {
			currStacks = append(currStacks, s)
		}
	}

	switch len(currStacks) {
	case 0:
		return Stack{}, errors.New("unable to infer current stack")
	case 1:
		return currStacks[0], nil
	default:
		var names []string
		for _, c := range currStacks {
			names = append(names, c.Name())
		}
		return Stack{}, fmt.Errorf("currently within multiple stacks: %s", strings.Join(names, ", "))
	}
}

type InferenceResult struct {
	InferredStacks  []Stack
	InferenceErrors []error
}

// InferStacks tries to infer all commit stacks from the log and returns an inference result
// containing the stacks it was able to infer and inference errors it encountered.
// If Infer encounters other kinds of errors, it will return (Inference{}, error)
func InferStacks(git Git, log libgit.Log) (InferenceResult, error) {
	graph, err := commitgraph.Compute(log)
	if err != nil {
		return InferenceResult{}, err
	}

	var allStacks []Stack
	var inferenceErrs []error
	for _, n := range graph.Nodes {
		if !n.IsSource() {
			continue
		}

		stacks, err := inferStacks(git, graph, n, Stack{}, 500)
		if err != nil {
			inferenceErrs = append(inferenceErrs, err)
			continue
		}

		// Validate the stacks
		if len(stacks) != 1 {
			for i, s := range stacks {
				var otherStackNames []string
				for j, s := range stacks {
					if i == j {
						continue
					}
					otherStackNames = append(otherStackNames, s.Name())
				}

				s.ValidationErrors = append(s.ValidationErrors, DivergenceError{
					StackName:       s.Name(),
					OtherStackNames: otherStackNames,
				})
				stacks[i] = s
			}
		}
		for i, s := range stacks {
			for _, c := range s.Commits {
				if len(c.LocalBranches) > 1 {
					s.ValidationErrors = append(s.ValidationErrors, BranchCollisionError{
						StackName: s.Name(),
						Branches:  c.LocalBranches,
					})
					stacks[i] = s
				}
			}
		}

		allStacks = append(allStacks, stacks...)
	}

	sortStacks(allStacks)
	return InferenceResult{
		InferredStacks:  allStacks,
		InferenceErrors: inferenceErrs,
	}, nil
}

func sortStacks(stacks []Stack) {
	slices.SortFunc(stacks, func(a, b Stack) int {
		if a.Name() < b.Name() {
			return -1
		} else if a.Name() == b.Name() {
			return 0
		} else {
			return 1
		}
	})
}

// inferStacks returns all stacks that start from currNode. This will never return an empty slice unless it's returning an error.
// An error is returned if inferStacks hits a terminal condition that prevents it from continuing.
func inferStacks(git Git, graph commitgraph.DAG, currNode commitgraph.Node, prevStack Stack, remainingRecursionDepth int) ([]Stack, error) {
	if remainingRecursionDepth < 0 {
		return nil, fmt.Errorf("failed to infer stacks, exceeded max recursion depth")
	}

	currStack, err := appendToStack(git, currNode, prevStack)
	if err != nil {
		return nil, err
	}

	// Base case
	if len(currNode.Children) == 0 {
		// Order from top to bottom
		slices.Reverse(currStack.Commits)
		return []Stack{currStack}, nil
	}

	var stacks []Stack
	for c := range currNode.Children {
		nextNode := graph.Nodes[c]
		got, err := inferStacks(git, graph, nextNode, currStack, remainingRecursionDepth-1)
		if err != nil {
			return nil, err
		}
		stacks = append(stacks, got...)
	}
	return stacks, nil
}

func appendToStack(git Git, currNode commitgraph.Node, prevStack Stack) (Stack, error) {
	slog.Debug("appendToStack", "currNode", currNode.Hash, "parents", currNode.Parents)

	if len(currNode.Parents) > 1 {
		got, err := git.GetBranchesContainingCommit(currNode.Hash)
		if err != nil {
			fmt.Printf("error: failed to get branches containing merge commit: %s", currNode.Hash)
			return Stack{}, MergeCommitError{
				MergeCommitHash: currNode.Hash,
			}
		}

		return Stack{}, MergeCommitError{
			MergeCommitHash:    currNode.Hash,
			ContainingBranches: got,
		}
	}

	c := Commit{
		Author:        currNode.Author,
		Date:          currNode.Date,
		Subject:       currNode.Subject,
		Hash:          currNode.Hash,
		LocalBranches: currNode.LocalBranches,
	}
	sort.Strings(c.LocalBranches)
	return Stack{
		Commits: append(slices.Clone(prevStack.Commits), c),
	}, nil
}

func assert(condition bool, msg string) {
	if !condition {
		panic(fmt.Sprintf("Invariant violated: %s", msg))
	}
}
