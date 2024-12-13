package stackslib

import (
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"sort"
	"strings"

	"github.com/raymondji/git-stack/commitgraph"
)

type Git interface {
	GetCurrentBranch() (string, error)
	commitgraph.Git
}

type Branch struct {
	Current bool
	Name    string
}

type Commit struct {
	Author      string
	Subject     string
	Date        string
	Hash        string
	LocalBranch *Branch
}

type Stacks struct {
	Entries []Stack
	Errors  []error
}

type Stack struct {
	Commits []Commit // Ordered from top to bottom, guaranteed to not be empty
	Error   error
}

// Guaranteed to return at least one branch
func (s Stack) LocalBranches() []Branch {
	var branches []Branch
	for _, c := range s.Commits {
		if c.LocalBranch == nil {
			continue
		}
		branches = append(branches, *c.LocalBranch)
	}
	return branches
}

func (s Stack) Name() string {
	return s.LocalBranches()[0].Name
}

func (s Stack) Current() bool {
	for _, b := range s.LocalBranches() {
		if b.Current {
			return true
		}
	}
	return false
}

func Compute(git Git, defaultBranch string) (Stacks, error) {
	graph, err := commitgraph.Compute(git, defaultBranch)
	if err != nil {
		return Stacks{}, err
	}

	currBranch, err := git.GetCurrentBranch()
	if err != nil {
		return Stacks{}, err
	}

	var result Stacks
	for _, n := range graph.Nodes {
		if !n.IsSource() {
			continue
		}

		stacks, err := buildStacks(graph, currBranch, n, Stack{})
		if errors.Is(err, errNoBranchesInStack) {
			fmt.Println(err.Error())
			continue
		} else if err != nil {
			result.Errors = append(result.Errors, err)
			continue
		}

		if len(stacks) == 1 {
			result.Entries = append(result.Entries, stacks[0])
		} else {
			err := newSharedCommitError(stacks)
			var stacksWithErr []Stack
			for _, s := range stacks {
				s.Error = err
				stacksWithErr = append(stacksWithErr, s)
			}
			result.Errors = append(result.Errors, err)
			result.Entries = append(result.Entries, stacksWithErr...)
		}
	}

	sortStacks(result.Entries)
	return result, nil
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

var errNoBranchesInStack = errors.New("no branches")

// buildStacks returns all stacks that start from currNode. This will never return an empty slice unless it's returning an error.
func buildStacks(graph commitgraph.DAG, currBranch string, currNode commitgraph.Node, prevStack Stack) ([]Stack, error) {
	currStack, err := addNodeToStack(currBranch, currNode, prevStack)
	if err != nil {
		return nil, err
	}

	// Base case
	if len(currNode.Children) == 0 {
		if len(currStack.LocalBranches()) == 0 {
			return nil, fmt.Errorf("%w, last commit in stack has no branch tags. stack: %v",
				errNoBranchesInStack, currStack)
		}

		// Order from top to bottom
		slices.Reverse(currStack.Commits)
		return []Stack{currStack}, nil
	}

	var stacks []Stack
	for c := range currNode.Children {
		got, err := buildStacks(graph, currBranch, graph.Nodes[c], currStack)
		if err != nil {
			return nil, err
		}
		stacks = append(stacks, got...)
	}
	return stacks, nil
}

type SharedCommitError struct {
	StackNames []string
}

var _ error = SharedCommitError{}

func newSharedCommitError(stacks []Stack) SharedCommitError {
	var names []string
	for _, s := range stacks {
		names = append(names, s.Name())
	}
	sort.Strings(names)
	return SharedCommitError{
		StackNames: names,
	}
}

func (e SharedCommitError) Error() string {
	return fmt.Sprintf("stacks %s have diverged, please rebase", strings.Join(e.StackNames, ", "))
}

type DuplicateBranchesError struct {
	Branches []string
}

var _ error = DuplicateBranchesError{}

func (e DuplicateBranchesError) Error() string {
	return fmt.Sprintf(
		"branches %v point to the same commit, please deduplicate",
		strings.Join(e.Branches, ", "))
}

type MergeCommitError struct {
	MergeCommitHash string
	PartialStack    Stack
}

var _ error = MergeCommitError{}

func (e MergeCommitError) Error() string {
	if len(e.PartialStack.LocalBranches()) > 0 {
		var branches []string
		for i, b := range e.PartialStack.LocalBranches() {
			var suffix string
			if i == 0 {
				suffix = " (top)"
			}
			branches = append(branches, fmt.Sprintf("%s%s", b.Name, suffix))
		}
		return fmt.Sprintf("found merge commit %v in partial stack %v, please undo merge. hint: try `git reflog show %s`",
			e.MergeCommitHash, strings.Join(branches, " <- "), e.PartialStack.LocalBranches()[0].Name)
	}
	return fmt.Sprintf("found merge commit %v, please undo merge", e.MergeCommitHash)
}

func addNodeToStack(currBranch string, currNode commitgraph.Node, prevStack Stack) (Stack, error) {
	slog.Debug("addNodeToStack", "currBranch", currBranch, "currNode", currNode.Hash, "parents", currNode.Parents)

	currStack := Stack{}
	// Deep copy
	currStack.Commits = append(currStack.Commits, prevStack.Commits...)

	c := Commit{
		Author:  currNode.Author,
		Date:    currNode.Date,
		Subject: currNode.Subject,
		Hash:    currNode.Hash,
	}
	switch len(currNode.LocalBranches) {
	case 0:
		currStack.Commits = append(currStack.Commits, c)
	case 1:
		b := Branch{
			Name: currNode.LocalBranches[0],
		}
		if b.Name == currBranch {
			b.Current = true
		}
		c.LocalBranch = &b
		currStack.Commits = append(currStack.Commits, c)
	default:
		return Stack{}, DuplicateBranchesError{
			Branches: currNode.LocalBranches,
		}
	}

	if len(currNode.Parents) > 1 {
		slices.Reverse(currStack.Commits)
		return Stack{}, MergeCommitError{
			MergeCommitHash: currNode.Hash,
			PartialStack:    currStack,
		}
	}

	return currStack, nil
}

var ErrMultipleCurrentStacks = errors.New("multiple current stacks")

func (stacks Stacks) GetCurrent() (Stack, error) {
	var currStacks []Stack
	for _, s := range stacks.Entries {
		if s.Current() {
			currStacks = append(currStacks, s)
		}
	}

	switch len(currStacks) {
	case 0:
		return Stack{}, fmt.Errorf("not in a stack")
	case 1:
		return currStacks[0], nil
	default:
		var names []string
		for _, s := range currStacks {
			names = append(names, s.Name())
		}
		return Stack{}, fmt.Errorf("%w, currently within multiple stacks, %s",
			ErrMultipleCurrentStacks, strings.Join(names, ", "))
	}
}
