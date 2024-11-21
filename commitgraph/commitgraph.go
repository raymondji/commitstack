package commitgraph

import (
	"fmt"

	"github.com/raymondji/git-stacked/git"
	"github.com/raymondji/git-stacked/utils"
)

type Git interface {
	GetLocalBranches() ([]string, error)
	GetLog(rangeStart, rangeEnd string) (git.Log, error)
	GetHash(branch string) (string, error)
}

type Builder struct {
	git Git
}

func NewBuilder(git Git) Builder {
	return Builder{
		git: git,
	}
}

type Node struct {
	Hash          string
	LocalBranches []string
	// Keys are hashes
	Children map[string]struct{}
}

type Graph struct {
	RootHash string
	// Keys are hashes
	Nodes map[string]Node
}

func (b Builder) Build(rootBranch string) (Graph, error) {
	localBranches, err := b.git.GetLocalBranches()
	if err != nil {
		return Graph{}, err
	}

	var getLogFuncs []func() (git.Log, error)
	for _, lb := range localBranches {
		getLogFuncs = append(getLogFuncs, func() (git.Log, error) {
			return b.git.GetLog(rootBranch, lb)
		})
	}
	logs, err := utils.RunConcurrently(getLogFuncs)
	if err != nil {
		return Graph{}, err
	}

	rootHash, err := b.git.GetHash(rootBranch)
	if err != nil {
		return Graph{}, err
	}
	root := Node{
		Hash:          rootHash,
		LocalBranches: []string{rootBranch},
	}
	g := Graph{
		RootHash: rootHash,
		Nodes: map[string]Node{
			root.Hash: root,
		},
	}
	for _, log := range logs {
		addToGraph(g, log)
	}

	return g, nil
}

func addToGraph(graph Graph, log git.Log) error {
	for _, commit := range log.Commits {
		if _, ok := graph.Nodes[commit.Hash]; ok {
			continue
		}
		graph.Nodes[commit.Hash] = Node{
			Hash:          commit.Hash,
			LocalBranches: commit.LocalBranches,
			Children:      map[string]struct{}{},
		}
	}

	// At this point all the parent nodes should have been added
	for _, commit := range log.Commits {
		for _, ph := range commit.ParentHashes {
			if _, ok := graph.Nodes[ph]; !ok {
				return fmt.Errorf("did not find parent in graph, child hash: %v, parent hash: %v", commit.Hash, ph)
			}
			graph.Nodes[ph].Children[commit.Hash] = struct{}{}
		}
	}

	return nil
}
