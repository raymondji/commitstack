package commitgraph

import (
	"fmt"
	"slices"

	"github.com/raymondji/git-stack-cli/libgit"
)

type Node struct {
	Author  string
	Subject string
	Date    string
	Hash    string
	// Sorted in reverse lexicographic order
	LocalBranches []string
}

func IsSource(n Node, dag DAG) bool {
	parents := dag.ParentEdges[n.Hash]
	return len(parents) == 0
}

func IsSink(n Node, dag DAG) bool {
	children := dag.ChildrenEdges[n.Hash]
	return len(children) == 0
}

type DAG struct {
	// Keys are short commit hashes
	Nodes map[string]Node
	// Keys are children, values are parents
	ParentEdges map[string]map[string]struct{}
	// Keys are parents, values are children
	ChildrenEdges map[string]map[string]struct{}
}

func ensureInitEdges(edges map[string]map[string]struct{}, k string) {
	if _, ok := edges[k]; !ok {
		edges[k] = map[string]struct{}{}
	}
}

func Compute(log libgit.Log) (DAG, error) {
	dag := DAG{
		Nodes:         map[string]Node{},
		ParentEdges:   map[string]map[string]struct{}{},
		ChildrenEdges: map[string]map[string]struct{}{},
	}
	for _, commit := range log.Commits {
		if _, ok := dag.Nodes[commit.Hash]; ok {
			return DAG{}, fmt.Errorf("duplicate commit in log, hash: %v", commit.Hash)
		}
		lb := commit.LocalBranches
		slices.Sort(lb)
		slices.Reverse(lb)
		dag.Nodes[commit.Hash] = Node{
			Author:        commit.Author,
			Date:          commit.Date,
			Subject:       commit.Subject,
			Hash:          commit.Hash,
			LocalBranches: lb,
		}
		ensureInitEdges(dag.ParentEdges, commit.Hash)
		ensureInitEdges(dag.ChildrenEdges, commit.Hash)
	}

	for _, commit := range log.Commits {
		for _, ph := range commit.ParentHashes {
			if _, ok := dag.Nodes[ph]; ok {
				dag.ParentEdges[commit.Hash][ph] = struct{}{}
				dag.ChildrenEdges[ph][commit.Hash] = struct{}{}
			}
		}
	}

	return dag, nil
}
