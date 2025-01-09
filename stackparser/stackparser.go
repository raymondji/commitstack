package stackparser

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/raymondji/git-stack-cli/libgit"
	"github.com/raymondji/git-stack-cli/stackparser/commitgraph"
)

type Stack struct {
	Name    string
	Commits map[string]*Commit // Guaranteed to not be empty
}

type Commit struct {
	commitgraph.Node
	// Keys are stack names
	// Values are 0 for any commits without branches,
	// or 1+ for commmits with branches.
	// A score of N indicates this commit was reachable following parent edges
	// from a commit with score N+1.
	StackBranchScore map[string]int
}

type NoTotalOrderError struct {
	StackName string
	// Each entry has length two
	IncomparableBranchPairs [][]string
}

func (e NoTotalOrderError) Error() string {
	var msgs []string
	assert(len(e.IncomparableBranchPairs) > 0, "must have at least 1 branch pair")
	for _, pair := range e.IncomparableBranchPairs {
		assert(len(pair) == 2, "branch pairs must have length 2")
		msgs = append(msgs, fmt.Sprintf("branch %s does not contain %s, and vice versa", pair[0], pair[1]))
	}
	// Printing all of them might be too overwhelming
	return fmt.Sprintf("%s: %s", e.StackName, msgs[0])
}

func (s Stack) Branches() []string {
	var branches []string
	for _, c := range s.Commits {
		branches = append(branches, c.LocalBranches...)
	}
	slices.Sort(branches)
	slices.Reverse(branches)
	return branches
}

// TotalOrderedBranches returns all branches associated with all commits in the stack.
// Guaranteed to return at least one branch. If invalid, returns NoTotalOrderError.
func (s Stack) TotalOrderedBranches() ([]string, error) {
	var branchCommits []*Commit
	for _, c := range s.Commits {
		if len(c.LocalBranches) > 0 {
			branchCommits = append(branchCommits, c)
		}
	}

	name := s.Name
	var incomparableBranchPairs [][]string
	slices.SortFunc(branchCommits, func(a, b *Commit) int {
		scoreA, ok := a.StackBranchScore[name]
		assert(ok, "commit must be in stack")
		scoreB, ok := b.StackBranchScore[name]
		assert(ok, "commit must be in stack")

		if scoreA < scoreB {
			return -1
		} else if scoreA > scoreB {
			return 1
		} else {
			// Arbitarily picking the first branch for now
			incomparableBranchPairs = append(incomparableBranchPairs,
				[]string{
					a.LocalBranches[0],
					b.LocalBranches[0],
				},
			)
			return 0
		}
	})
	if len(incomparableBranchPairs) > 0 {
		return nil, NoTotalOrderError{
			StackName:               s.Name,
			IncomparableBranchPairs: incomparableBranchPairs,
		}
	}

	var branches []string
	for _, c := range branchCommits {
		branches = append(branches, c.LocalBranches...)
	}
	return branches, nil
}

func (s Stack) DivergesFrom() map[string]struct{} {
	name := s.Name
	divergesFrom := map[string]struct{}{}
	for _, c := range s.Commits {
		for stackName := range c.StackBranchScore {
			if stackName == name {
				continue
			}
			divergesFrom[stackName] = struct{}{}
		}
	}

	return divergesFrom
}

func (s Stack) IsCurrent(currCommit string) bool {
	for _, c := range s.Commits {
		if c.Hash == currCommit {
			return true
		}
	}
	return false
}

func GetCurrent(stacks []Stack, currCommit string) (Stack, error) {
	var currStacks []Stack
	for _, s := range stacks {
		if s.IsCurrent(currCommit) {
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
		for _, s := range currStacks {
			names = append(names, s.Name)
		}
		// TODO: add message to indicate you can rerun with `git stack [cmd] [stack]` to disambiguate
		return Stack{}, fmt.Errorf("currently within multiple stacks: %s", strings.Join(names, ", "))
	}
}

type dag struct {
	commits     map[string]*Commit
	parentEdges map[string]map[string]struct{}
}

// ParseStacks parses commit stacks from the git commit log
func ParseStacks(log libgit.Log) ([]Stack, error) {
	rawGraph, err := commitgraph.Compute(log)
	if err != nil {
		return nil, err
	}

	graph := dag{
		parentEdges: rawGraph.ParentEdges,
		commits:     map[string]*Commit{},
	}
	for k, n := range rawGraph.Nodes {
		graph.commits[k] = &Commit{
			Node:             n,
			StackBranchScore: map[string]int{},
		}
	}

	var stacks []Stack
	for k, c := range graph.commits {
		if !commitgraph.IsSink(c.Node, rawGraph) {
			continue
		}
		assert(len(c.LocalBranches) > 0, "no branches pointing to leaf commit")

		stack := Stack{
			Name:    c.LocalBranches[0],
			Commits: map[string]*Commit{},
		}
		// TODO: Can we adapt djikstra's to reduce the time complexity?
		// Although this is multi-source and multi-sink as-is.
		// Can we transform this into a single-source problem
		// by making `defaultBranch` the only source
		// instead of starting from the leaf commits?
		err := parseStack(k, 1, graph, stack, 250)
		if err != nil {
			return nil, err
		}
		stacks = append(stacks, stack)
	}
	sortStacks(stacks)

	return stacks, nil
}

func sortStacks(stacks []Stack) {
	slices.SortFunc(stacks, func(a, b Stack) int {
		if a.Name < b.Name {
			return -1
		} else if a.Name == b.Name {
			return 0
		} else {
			return 1
		}
	})
}

func parseStack(
	nodeKey string, branchCommitScore int,
	graph dag, stack Stack,
	remainingRecursionDepth int,
) error {
	if remainingRecursionDepth < 0 {
		return fmt.Errorf("failed to infer stack, exceeded max recursion depth")
	}

	commit := graph.commits[nodeKey]
	stack.Commits[commit.Hash] = commit

	if len(commit.LocalBranches) == 0 {
		commit.StackBranchScore[stack.Name] = 0
	} else {
		commit.StackBranchScore[stack.Name] = max(
			commit.StackBranchScore[stack.Name],
			branchCommitScore,
		)
		branchCommitScore = commit.StackBranchScore[stack.Name] + 1
	}

	for nextKey := range graph.parentEdges[nodeKey] {
		err := parseStack(nextKey, branchCommitScore, graph, stack, remainingRecursionDepth-1)
		if err != nil {
			return err
		}
	}
	return nil
}

func assert(condition bool, msg string) {
	if !condition {
		panic(fmt.Sprintf("Invariant violated: %s", msg))
	}
}
