package commitgraph_test

import (
	"testing"

	"github.com/raymondji/git-stack-cli/inference/commitgraph"
	"github.com/raymondji/git-stack-cli/libgit"
	"github.com/stretchr/testify/require"
)

func TestCompute(t *testing.T) {
	cases := map[string]struct {
		log  libgit.Log
		want commitgraph.DAG
	}{
		"empty": {
			log: libgit.Log{
				Commits: nil,
			},
			want: commitgraph.DAG{
				Nodes:         map[string]commitgraph.Node{},
				ParentEdges:   map[string]map[string]struct{}{},
				ChildrenEdges: map[string]map[string]struct{}{},
			},
		},
		"one": {
			log: libgit.Log{
				Commits: []libgit.Commit{
					{
						Hash:          "c1",
						ParentHashes:  []string{"p1"},
						LocalBranches: []string{"b1"},
					},
				},
			},
			want: commitgraph.DAG{
				Nodes: map[string]commitgraph.Node{
					"c1": {
						Hash:          "c1",
						LocalBranches: []string{"b1"},
					},
				},
				ParentEdges: map[string]map[string]struct{}{
					"c1": {},
				},
				ChildrenEdges: map[string]map[string]struct{}{
					"c1": {},
				},
			},
		},
		"sorts branches": {
			log: libgit.Log{
				Commits: []libgit.Commit{
					{
						Hash:          "c1",
						ParentHashes:  []string{"p1"},
						LocalBranches: []string{"b", "b1"},
					},
				},
			},
			want: commitgraph.DAG{
				Nodes: map[string]commitgraph.Node{
					"c1": {
						Hash:          "c1",
						LocalBranches: []string{"b1", "b"},
					},
				},
				ParentEdges: map[string]map[string]struct{}{
					"c1": {},
				},
				ChildrenEdges: map[string]map[string]struct{}{
					"c1": {},
				},
			},
		},
		"single parent": {
			log: libgit.Log{
				Commits: []libgit.Commit{
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
			want: commitgraph.DAG{
				Nodes: map[string]commitgraph.Node{
					"c1": {
						Hash:          "c1",
						LocalBranches: []string{"feat/pt2"},
					},
					"p1": {
						Hash:          "p1",
						LocalBranches: []string{"feat/pt1"},
					},
				},
				ParentEdges: map[string]map[string]struct{}{
					"c1": {
						"p1": struct{}{},
					},
					"p1": {},
				},
				ChildrenEdges: map[string]map[string]struct{}{
					"p1": {
						"c1": struct{}{},
					},
					"c1": {},
				},
			},
		},
		"multiple parents": {
			log: libgit.Log{
				Commits: []libgit.Commit{
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
			want: commitgraph.DAG{
				Nodes: map[string]commitgraph.Node{
					"c3": {
						Hash:          "c3",
						LocalBranches: []string{"feat/pt3"},
					},
					"c2": {
						Hash:          "c2",
						LocalBranches: []string{"feat/pt2"},
					},
					"c1": {
						Hash:          "c1",
						LocalBranches: []string{"feat/pt1"},
					},
				},
				ParentEdges: map[string]map[string]struct{}{
					"c3": {
						"c2": struct{}{},
						"c1": struct{}{},
					},
					"c2": {
						"c1": struct{}{},
					},
					"c1": {},
				},
				ChildrenEdges: map[string]map[string]struct{}{
					"c3": {},
					"c2": {
						"c3": struct{}{},
					},
					"c1": {
						"c2": struct{}{},
						"c3": struct{}{},
					},
				},
			},
		},
		"multiple sources": {
			log: libgit.Log{
				Commits: []libgit.Commit{
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
			want: commitgraph.DAG{
				Nodes: map[string]commitgraph.Node{
					"c3": {
						Hash:          "c3",
						LocalBranches: []string{"featA/pt2"},
					},
					"c2": {
						Hash:          "c2",
						LocalBranches: []string{"featA/pt1"},
					},
					"c1": {
						Hash:          "c1",
						LocalBranches: []string{"featB/pt1"},
					},
				},
				ParentEdges: map[string]map[string]struct{}{
					"c3": {
						"c2": struct{}{},
					},
					"c2": {},
					"c1": {},
				},
				ChildrenEdges: map[string]map[string]struct{}{
					"c3": {},
					"c2": {
						"c3": struct{}{},
					},
					"c1": {},
				},
			},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			got, err := commitgraph.Compute(c.log)
			require.NoError(t, err)
			require.Equal(t, c.want, got)
		})
	}
}
