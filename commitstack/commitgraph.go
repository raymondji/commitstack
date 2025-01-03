package commitstack

import (
	"fmt"
)

type node struct {
	Author        string
	Subject       string
	Date          string
	Hash          string
	LocalBranches []string
	// Keys are hashes. All keys must be present in the DirectedAcyclicGraph.
	Children map[string]struct{}
	// Keys are hashes. All keys must be present in the DirectedAcyclicGraph.
	Parents map[string]struct{}
}

func (n node) IsSource() bool {
	return len(n.Parents) == 0
}

type directedAcyclicGraph struct {
	// Keys are hashes
	Nodes map[string]node
}

func computeDAG(git Git, defaultBranch string) (directedAcyclicGraph, error) {
	log, err := git.LogAll(defaultBranch)
	if err != nil {
		return directedAcyclicGraph{}, err
	}

	dag := directedAcyclicGraph{
		Nodes: map[string]node{},
	}
	for _, commit := range log.Commits {
		if _, ok := dag.Nodes[commit.Hash]; ok {
			return directedAcyclicGraph{}, fmt.Errorf("duplicate commit in log, hash: %v", commit.Hash)
		}
		dag.Nodes[commit.Hash] = node{
			Author:        commit.Author,
			Date:          commit.Date,
			Subject:       commit.Subject,
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
