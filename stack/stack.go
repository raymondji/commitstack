package stack

import (
	"fmt"

	"github.com/raymondji/git-stacked/commitgraph"
)

type Github interface {
}

type Gitlab interface {
}

type Git interface {
	GetCurrentBranch() (string, error)
	commitgraph.Git
}

type Stack struct {
	Current       bool
	LocalBranches []Branch
}

type Branch struct {
	Current bool
	Name    string
}

func GetCurrent(git Git, rootBranch string) (Stack, error) {
	stacks, err := GetAll(git, rootBranch)
	if err != nil {
		return Stack{}, err
	}

	for _, s := range stacks {
		if s.Current {
			return s, nil
		}
	}

	return Stack{}, fmt.Errorf("not currently in a stack")
}

func GetAll(git Git, rootBranch string) ([]Stack, error) {
	graph, err := commitgraph.NewBuilder(git).Build(rootBranch)
	if err != nil {
		return nil, err
	}

	currBranch, err := git.GetCurrentBranch()
	if err != nil {
		return nil, err
	}

	root := graph.Nodes[graph.RootHash]
	var stacks []Stack
	for rootChild := range root.Children {
		stack, err := buildStack(graph, graph.Nodes[rootChild], currBranch)
		if err != nil {
			return nil, err
		}
		stacks = append(stacks, stack)
	}

	return stacks, nil
}

func buildStack(
	graph commitgraph.Graph, rootChild commitgraph.Node, currBranch string,
) (Stack, error) {
	var stack Stack
	curr := rootChild

outer:
	for {
		switch len(curr.LocalBranches) {
		case 0:
			// do nothing
		case 1:
			b := Branch{
				Name: curr.LocalBranches[0],
			}
			if b.Name == currBranch {
				b.Current = true
				stack.Current = true
			}
			stack.LocalBranches = append(stack.LocalBranches, b)
		default:
			return Stack{}, fmt.Errorf(
				"invalid git commit graph, commit %s is referenced by multiple local branches: %v",
				curr.Hash, curr.LocalBranches)
		}

		switch len(curr.Children) {
		case 0:
			break outer
		case 1:
			for c := range curr.Children {
				curr = graph.Nodes[c]
			}
		case 2:
			return Stack{}, fmt.Errorf(
				"invalid git commit graph, commit %s has multiple children: %v",
				curr.Hash, curr.Children)
		}
	}

	return stack, nil
}
