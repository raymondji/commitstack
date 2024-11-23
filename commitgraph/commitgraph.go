package commitgraph

import (
	"fmt"

	"github.com/raymondji/git-stacked/gitlib"
)

type Git interface {
	LogAll(notReachableFrom string) (gitlib.Log, error)
}

type Node struct {
	Hash          string
	LocalBranches []string
	// Keys are hashes. All keys must be present in the DAG.
	Children map[string]struct{}
	// Keys are hashes. All keys must be present in the DAG.
	Parents map[string]struct{}
}

func (n Node) IsSource() bool {
	return len(n.Parents) == 0
}

type DAG struct {
	// Keys are hashes
	Nodes map[string]Node
}

func Build(git Git, defaultBranch string) (DAG, error) {
	log, err := git.LogAll(defaultBranch)
	if err != nil {
		return DAG{}, err
	}
	fmt.Println("got log")
	fmt.Println(log)

	dag := DAG{
		Nodes: map[string]Node{},
	}
	for _, commit := range log.Commits {
		if _, ok := dag.Nodes[commit.Hash]; ok {
			return DAG{}, fmt.Errorf("duplicate commit in log, hash: %v", commit.Hash)
		}
		dag.Nodes[commit.Hash] = Node{
			Hash:          commit.Hash,
			LocalBranches: commit.LocalBranches,
			Children:      map[string]struct{}{},
			Parents:       map[string]struct{}{},
		}
	}
	for _, commit := range log.Commits {
		for _, ph := range commit.ParentHashes {
			if _, ok := dag.Nodes[ph]; ok {
				dag.Nodes[commit.Hash].Parents[ph] = struct{}{}
				dag.Nodes[ph].Children[commit.Hash] = struct{}{}
			}
		}
	}

	return dag, nil
}
