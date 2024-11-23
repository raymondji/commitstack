package stack

import (
	"errors"
	"fmt"
	"slices"
	"strings"

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
	Name          string // This is set to the top branch in the stack
	Current       bool
	LocalBranches []Branch // Ordered from top to bottom
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

	return Stack{}, fmt.Errorf("not in a stack")
}

func GetAll(git Git, defaultBranch string) ([]Stack, error) {
	graph, err := commitgraph.Build(git, defaultBranch)
	if err != nil {
		return nil, err
	}

	currBranch, err := git.GetCurrentBranch()
	if err != nil {
		return nil, err
	}

	var stacks []Stack
	for _, n := range graph.Nodes {
		if !n.IsSource() {
			continue
		}

		stack, err := buildStack(graph, n, currBranch)
		if errors.Is(err, errNoBranchesInStack) {
			fmt.Println(err.Error())
			continue
		} else if err != nil {
			return nil, err
		}
		stacks = append(stacks, stack)
	}

	slices.SortFunc(stacks, func(a, b Stack) int {
		if a.Name < b.Name {
			return -1
		} else if a.Name == b.Name {
			return 0
		} else {
			return 1
		}
	})
	return stacks, nil
}

var errNoBranchesInStack = errors.New("no branches")

func buildStack(
	graph commitgraph.DAG, source commitgraph.Node, currBranch string,
) (Stack, error) {
	var stack Stack
	curr := source

outer:
	for {
		if len(curr.Parents) > 1 {
			var parents []string
			for p := range curr.Parents {
				parents = append(parents, p)
			}
			return Stack{}, fmt.Errorf("unsupported git commit graph, commit %s has multiple parents: %v",
				curr.Hash, strings.Join(parents, ", "))
		}

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
				"unsupported commit graph, commit %s is referenced by multiple local branches: %v",
				curr.Hash, strings.Join(curr.LocalBranches, ", "))
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

	if len(stack.LocalBranches) == 0 {
		return Stack{}, fmt.Errorf("unexpected source node, no branch tags in stack: %v, err: %w", source.Hash, errNoBranchesInStack)
	}

	slices.Reverse(stack.LocalBranches)
	stack.Name = stack.LocalBranches[0].Name
	return stack, nil
}
