package inference_test

import (
	"testing"

	"github.com/raymondji/git-stack-cli/inference"
	"github.com/raymondji/git-stack-cli/inference/commitgraph"
	"github.com/raymondji/git-stack-cli/libgit"
	"github.com/stretchr/testify/require"
)

func TestInferStacks(t *testing.T) {
	cases := map[string]struct {
		log                            libgit.Log
		commitHashToContainingBranches map[string][]string
		want                           []inference.Stack
		// Keys are stack names
		wantTotalOrderedBranches    map[string][]string
		wantTotalOrderedBranchesErr map[string]struct{}
	}{
		"empty": {
			log: libgit.Log{
				Commits: nil,
			},
			want: nil,
		},
		"one": {
			log: libgit.Log{
				Commits: []libgit.Commit{
					{
						Hash:          "c1",
						ParentHashes:  []string{"p1"},
						LocalBranches: []string{"dev"},
					},
				},
			},
			want: []inference.Stack{
				{
					Name: "dev",
					Commits: map[string]*inference.Commit{
						"c1": {
							Node: commitgraph.Node{
								Hash:          "c1",
								LocalBranches: []string{"dev"},
							},
							StackBranchScore: map[string]int{
								"dev": 1,
							},
						},
					},
				},
			},
			wantTotalOrderedBranches: map[string][]string{
				"dev": {"dev"},
			},
		},
		"commit with multiple branches": {
			log: libgit.Log{
				Commits: []libgit.Commit{
					{
						Hash:          "c1",
						ParentHashes:  []string{"p1"},
						LocalBranches: []string{"dev", "dev2"},
					},
				},
			},
			want: []inference.Stack{
				{
					Name: "dev2",
					Commits: map[string]*inference.Commit{
						"c1": {
							Node: commitgraph.Node{
								Hash:          "c1",
								LocalBranches: []string{"dev2", "dev"},
							},
							StackBranchScore: map[string]int{
								"dev2": 1,
							},
						},
					},
				},
			},
			wantTotalOrderedBranches: map[string][]string{
				"dev2": {"dev2", "dev"},
			},
		},
		"single parent": {
			log: libgit.Log{
				Commits: []libgit.Commit{
					{
						Hash:          "c2",
						ParentHashes:  []string{"c1"},
						LocalBranches: []string{"feat/pt3"},
					},
					{
						Hash:          "c1",
						ParentHashes:  []string{"c0"},
						LocalBranches: []string{"feat/pt1", "feat/pt2"},
					},
				},
			},
			want: []inference.Stack{
				{
					Name: "feat/pt3",
					Commits: map[string]*inference.Commit{
						"c2": {
							Node: commitgraph.Node{
								Hash:          "c2",
								LocalBranches: []string{"feat/pt3"},
							},
							StackBranchScore: map[string]int{
								"feat/pt3": 1,
							},
						},
						"c1": {
							Node: commitgraph.Node{
								Hash:          "c1",
								LocalBranches: []string{"feat/pt2", "feat/pt1"},
							},
							StackBranchScore: map[string]int{
								"feat/pt3": 2,
							},
						},
					},
				},
			},
			wantTotalOrderedBranches: map[string][]string{
				"feat/pt3": {"feat/pt3", "feat/pt2", "feat/pt1"},
			},
		},
		"multiple parents with no total order for branches": {
			log: libgit.Log{
				Commits: []libgit.Commit{
					{
						Hash:          "c3",
						ParentHashes:  []string{"c2", "c1"},
						LocalBranches: []string{"featC"},
					},
					{
						Hash:          "c2",
						ParentHashes:  []string{"c0"},
						LocalBranches: []string{"featA"},
					},
					{
						Hash:          "c1",
						ParentHashes:  []string{"c0"},
						LocalBranches: []string{"featB"},
					},
				},
			},
			want: []inference.Stack{
				{
					Name: "featC",
					Commits: map[string]*inference.Commit{
						"c3": {
							Node: commitgraph.Node{
								Hash:          "c3",
								LocalBranches: []string{"featC"},
							},
							StackBranchScore: map[string]int{
								"featC": 1,
							},
						},
						"c2": {
							Node: commitgraph.Node{
								Hash:          "c2",
								LocalBranches: []string{"featA"},
							},
							StackBranchScore: map[string]int{
								"featC": 2,
							},
						},
						"c1": {
							Node: commitgraph.Node{
								Hash:          "c1",
								LocalBranches: []string{"featB"},
							},
							StackBranchScore: map[string]int{
								"featC": 2,
							},
						},
					},
				},
			},
			wantTotalOrderedBranchesErr: map[string]struct{}{
				"featC": {},
			},
		},
		"multiple parents with valid total order for branches": {
			log: libgit.Log{
				Commits: []libgit.Commit{
					{
						Hash:          "c6",
						ParentHashes:  []string{"c5", "c4"},
						LocalBranches: []string{"feat3"},
					},
					{
						Hash:         "c5",
						ParentHashes: []string{"c3"},
					},
					{
						Hash:          "c4",
						ParentHashes:  []string{"c3", "c2"},
						LocalBranches: []string{"feat2"},
					},
					{
						Hash:         "c3",
						ParentHashes: []string{"c1"},
					},
					{
						Hash:          "c2",
						ParentHashes:  []string{"c1"},
						LocalBranches: []string{"feat1"},
					},
					{
						Hash:         "c1",
						ParentHashes: []string{"c0"},
					},
				},
			},
			want: []inference.Stack{
				{
					Name: "feat3",
					Commits: map[string]*inference.Commit{
						"c6": {
							Node: commitgraph.Node{
								Hash:          "c6",
								LocalBranches: []string{"feat3"},
							},
							StackBranchScore: map[string]int{
								"feat3": 1,
							},
						},
						"c5": {
							Node: commitgraph.Node{
								Hash: "c5",
							},
							StackBranchScore: map[string]int{},
						},
						"c4": {
							Node: commitgraph.Node{
								Hash:          "c4",
								LocalBranches: []string{"feat2"},
							},
							StackBranchScore: map[string]int{
								"feat3": 2,
							},
						},
						"c3": {
							Node: commitgraph.Node{
								Hash: "c3",
							},
							StackBranchScore: map[string]int{},
						},
						"c2": {
							Node: commitgraph.Node{
								Hash:          "c2",
								LocalBranches: []string{"feat1"},
							},
							StackBranchScore: map[string]int{
								"feat3": 3,
							},
						},
						"c1": {
							Node: commitgraph.Node{
								Hash: "c1",
							},
							StackBranchScore: map[string]int{},
						},
					},
				},
			},
			wantTotalOrderedBranches: map[string][]string{
				"feat3": {"feat3", "feat2", "feat1"},
			},
		},
		"multiple stacks": {
			log: libgit.Log{
				Commits: []libgit.Commit{
					{
						Hash:          "c4",
						ParentHashes:  []string{"c3"},
						LocalBranches: []string{"featA/pt2"},
					},
					{
						Hash:          "c3",
						ParentHashes:  []string{"c2"},
						LocalBranches: nil,
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
			want: []inference.Stack{
				{
					Name: "featA/pt2",
					Commits: map[string]*inference.Commit{
						"c4": {
							Node: commitgraph.Node{
								Hash:          "c4",
								LocalBranches: []string{"featA/pt2"},
							},
							StackBranchScore: map[string]int{
								"featA/pt2": 1,
							},
						},
						"c3": {
							Node: commitgraph.Node{
								Hash: "c3",
							},
							StackBranchScore: map[string]int{},
						},
						"c2": {
							Node: commitgraph.Node{
								Hash:          "c2",
								LocalBranches: []string{"featA/pt1"},
							},
							StackBranchScore: map[string]int{
								"featA/pt2": 2,
							},
						},
					},
				},
				{
					Name: "featB/pt1",
					Commits: map[string]*inference.Commit{
						"c1": {
							Node: commitgraph.Node{
								Hash:          "c1",
								LocalBranches: []string{"featB/pt1"},
							},
							StackBranchScore: map[string]int{
								"featB/pt1": 1,
							},
						},
					},
				},
			},
			wantTotalOrderedBranches: map[string][]string{
				"featA/pt2": {"featA/pt2", "featA/pt1"},
				"featB/pt1": {"featB/pt1"},
			},
		},
		"divergent stacks": {
			log: libgit.Log{
				Commits: []libgit.Commit{
					{
						Hash:          "c3",
						ParentHashes:  []string{"c1"},
						LocalBranches: []string{"featC"},
					},
					{
						Hash:          "c2",
						ParentHashes:  []string{"c1"},
						LocalBranches: []string{"featB"},
					},
					{
						Hash:          "c1",
						ParentHashes:  []string{"c0"},
						LocalBranches: []string{"featA"},
					},
				},
			},
			want: []inference.Stack{
				{
					Name: "featB",
					Commits: map[string]*inference.Commit{
						"c2": {
							Node: commitgraph.Node{
								Hash:          "c2",
								LocalBranches: []string{"featB"},
							},
							StackBranchScore: map[string]int{
								"featB": 1,
							},
						},
						"c1": {
							Node: commitgraph.Node{
								Hash:          "c1",
								LocalBranches: []string{"featA"},
							},
							StackBranchScore: map[string]int{
								"featC": 2,
								"featB": 2,
							},
						},
					},
				},
				{
					Name: "featC",
					Commits: map[string]*inference.Commit{
						"c3": {
							Node: commitgraph.Node{
								Hash:          "c3",
								LocalBranches: []string{"featC"},
							},
							StackBranchScore: map[string]int{
								"featC": 1,
							},
						},
						"c1": {
							Node: commitgraph.Node{
								Hash:          "c1",
								LocalBranches: []string{"featA"},
							},
							StackBranchScore: map[string]int{
								"featC": 2,
								"featB": 2,
							},
						},
					},
				},
			},
			wantTotalOrderedBranches: map[string][]string{
				"featC": {"featC", "featA"},
				"featB": {"featB", "featA"},
			},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			got, err := inference.InferStacks(c.log)
			require.NoError(t, err)
			require.Equal(t, c.want, got)

			stacksMap := map[string]inference.Stack{}
			for _, s := range got {
				stacksMap[s.Name] = s
			}
			for sName, branches := range c.wantTotalOrderedBranches {
				s, ok := stacksMap[sName]
				require.True(t, ok)
				gotBranches, err := s.TotalOrderedBranches()
				require.NoError(t, err)
				require.Equal(t, branches, gotBranches)
			}
			for sName := range c.wantTotalOrderedBranchesErr {
				s, ok := stacksMap[sName]
				require.True(t, ok)
				_, err := s.TotalOrderedBranches()
				require.Error(t, err)
			}
		})
	}
}
