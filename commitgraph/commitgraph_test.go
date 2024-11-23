package commitgraph_test

import (
	"testing"

	"github.com/raymondji/git-stacked/commitgraph"
	"github.com/raymondji/git-stacked/gitlib"
	"github.com/stretchr/testify/require"
)

func TestBuild(t *testing.T) {
	cases := map[string]struct {
		log  gitlib.Log
		want commitgraph.DAG
	}{
		"empty": {
			log: gitlib.Log{
				Commits: nil,
			},
			want: toDAG(),
		},
		"one": {
			log: gitlib.Log{
				Commits: []gitlib.Commit{
					{
						Hash:          "c1",
						ParentHashes:  []string{"p1"},
						LocalBranches: []string{"b1"},
					},
				},
			},
			want: toDAG(
				commitgraph.Node{
					Hash:          "c1",
					Children:      map[string]struct{}{},
					Parents:       map[string]struct{}{},
					LocalBranches: []string{"b1"},
				}),
		},
		"with single parent": {
			log: gitlib.Log{
				Commits: []gitlib.Commit{
					{
						Hash:          "c1",
						ParentHashes:  []string{"p1"},
						LocalBranches: []string{"feat/pt2"},
					},
					{
						Hash:          "p1",
						ParentHashes:  []string{"p2"},
						LocalBranches: []string{"feat/pt1"},
					},
				},
			},
			want: toDAG(
				commitgraph.Node{
					Hash:     "c1",
					Children: map[string]struct{}{},
					Parents: map[string]struct{}{
						"p1": {},
					},
					LocalBranches: []string{"feat/pt2"},
				},
				commitgraph.Node{
					Hash: "p1",
					Children: map[string]struct{}{
						"c1": {},
					},
					Parents:       map[string]struct{}{},
					LocalBranches: []string{"feat/pt1"},
				},
			),
		},
		"with multiple parents": {
			log: gitlib.Log{
				Commits: []gitlib.Commit{
					{
						Hash:          "c3",
						ParentHashes:  []string{"c2", "c1"},
						LocalBranches: []string{"feat/pt3"},
					},
					{
						Hash:          "c2",
						ParentHashes:  []string{"c1"},
						LocalBranches: []string{"feat/pt2"},
					},
					{
						Hash:          "c1",
						ParentHashes:  []string{},
						LocalBranches: []string{"feat/pt1"},
					},
				},
			},
			want: toDAG(
				commitgraph.Node{
					Hash:     "c3",
					Children: map[string]struct{}{},
					Parents: map[string]struct{}{
						"c1": {},
						"c2": {},
					},
					LocalBranches: []string{"feat/pt3"},
				},
				commitgraph.Node{
					Hash: "c2",
					Children: map[string]struct{}{
						"c3": {},
					},
					Parents: map[string]struct{}{
						"c1": {},
					},
					LocalBranches: []string{"feat/pt2"},
				},
				commitgraph.Node{
					Hash: "c1",
					Children: map[string]struct{}{
						"c2": {},
						"c3": {},
					},
					Parents:       map[string]struct{}{},
					LocalBranches: []string{"feat/pt1"},
				},
			),
		},
		"multiple sources": {
			log: gitlib.Log{
				Commits: []gitlib.Commit{
					{
						Hash:          "c3",
						ParentHashes:  []string{"c2"},
						LocalBranches: []string{"featA/pt2"},
					},
					{
						Hash:          "c2",
						ParentHashes:  []string{"c0"},
						LocalBranches: []string{"featA/pt1"},
					},
					{
						Hash:          "c1",
						ParentHashes:  []string{"c0"},
						LocalBranches: []string{"featB/pt1"},
					},
				},
			},
			want: toDAG(
				commitgraph.Node{
					Hash:     "c3",
					Children: map[string]struct{}{},
					Parents: map[string]struct{}{
						"c2": {},
					},
					LocalBranches: []string{"featA/pt2"},
				},
				commitgraph.Node{
					Hash: "c2",
					Children: map[string]struct{}{
						"c3": {},
					},
					Parents:       map[string]struct{}{},
					LocalBranches: []string{"featA/pt1"},
				},
				commitgraph.Node{
					Hash:          "c1",
					Children:      map[string]struct{}{},
					LocalBranches: []string{"featB/pt1"},
					Parents:       map[string]struct{}{},
				},
			),
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			fg := &FakeGit{
				Log: c.log,
			}

			got, err := commitgraph.Build(fg, "main")
			require.NoError(t, err)
			require.Equal(t, c.want, got)
		})
	}
}

func toDAG(nodes ...commitgraph.Node) commitgraph.DAG {
	dag := commitgraph.DAG{
		Nodes: map[string]commitgraph.Node{},
	}
	for _, n := range nodes {
		dag.Nodes[n.Hash] = n
	}
	return dag
}

type FakeGit struct {
	Log gitlib.Log
}

func (fg *FakeGit) LogAll(notReachableFrom string) (gitlib.Log, error) {
	return fg.Log, nil
}
