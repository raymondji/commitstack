package stackslib

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/raymondji/git-stacked/commitgraph"
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
	Hash          string
	LocalBranches []string
}

type Stacks struct {
	Entries []Stack

	// Each entry is a list of stack names that share some number of commits
	SharingHistory [][]string
}

type Stack struct {
	Commits       []Commit // Ordered from top to bottom, guaranteed to not be empty
	LocalBranches []Branch // Ordered from top to bottom, guaranteed to not be empty
}

func (s Stack) Name() string {
	return s.LocalBranches[0].Name
}

func (s Stack) Current() bool {
	for _, b := range s.LocalBranches {
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
			return Stacks{}, err
		}

		if len(stacks) == 1 {
			result.Entries = append(result.Entries, stacks[0])
		} else {
			var names []string
			for _, s := range stacks {
				names = append(names, s.Name())
			}
			result.SharingHistory = append(result.SharingHistory, names)
			result.Entries = append(result.Entries, stacks...)
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
	if len(currNode.Parents) > 1 {
		var parents []string
		for p := range currNode.Parents {
			parents = append(parents, p)
		}
		return nil, fmt.Errorf("unsupported git commit graph, commit %s has multiple parents: %v",
			currNode.Hash, strings.Join(parents, ", "))
	}

	currStack, err := addNodeToStack(currBranch, currNode, prevStack)
	if err != nil {
		return nil, err
	}

	// Base case
	if len(currNode.Children) == 0 {
		if len(currStack.LocalBranches) == 0 {
			return nil, fmt.Errorf("%w, last commit in stack has no branch tags. stack: %v",
				errNoBranchesInStack, currStack)
		}

		// Order from top to bottom
		slices.Reverse(currStack.Commits)
		slices.Reverse(currStack.LocalBranches)
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

func addNodeToStack(currBranch string, currNode commitgraph.Node, prevStack Stack) (Stack, error) {
	currStack := Stack{}
	// Deep copy
	currStack.LocalBranches = append(currStack.LocalBranches, prevStack.LocalBranches...)
	currStack.Commits = append(currStack.Commits, prevStack.Commits...)
	currStack.Commits = append(currStack.Commits, Commit{
		Hash:          currNode.Hash,
		LocalBranches: currNode.LocalBranches,
	})

	switch len(currNode.LocalBranches) {
	case 0:
		return currStack, nil
	case 1:
		b := Branch{
			Name: currNode.LocalBranches[0],
		}
		if b.Name == currBranch {
			b.Current = true
		}
		currStack.LocalBranches = append(currStack.LocalBranches, b)
		return currStack, nil
	default:
		return Stack{}, fmt.Errorf(
			"unsupported commit graph, commit %s is referenced by multiple local branches: %v",
			currNode.Hash, strings.Join(currNode.LocalBranches, ", "))
	}
}

var ErrMultipleCurrentStacks = errors.New("multiple current stacks")

func ComputeCurrent(git Git, rootBranch string) (Stack, error) {
	stacks, err := Compute(git, rootBranch)
	if err != nil {
		return Stack{}, err
	}

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
