package stackslib_test

import (
	"testing"

	"github.com/raymondji/git-stacked/gitlib"
	"github.com/raymondji/git-stacked/stackslib"
	"github.com/stretchr/testify/require"
)

func TestCompute(t *testing.T) {
	cases := map[string]struct {
		currBranch      string
		log             gitlib.Log
		want            stackslib.Stacks
		wantErrContains string
	}{
		"empty": {
			log: gitlib.Log{
				Commits: nil,
			},
			currBranch: "main",
			want:       stackslib.Stacks{},
		},
		"one": {
			currBranch: "dev",
			log: gitlib.Log{
				Commits: []gitlib.Commit{
					{
						Hash:          "c1",
						ParentHashes:  []string{"p1"},
						LocalBranches: []string{"dev"},
					},
				},
			},
			want: stackslib.Stacks{
				Entries: []stackslib.Stack{
					{
						Commits: []stackslib.Commit{
							{
								Hash:          "c1",
								LocalBranches: []string{"dev"},
							},
						},
						LocalBranches: []stackslib.Branch{
							{
								Current: true,
								Name:    "dev",
							},
						},
					},
				},
			},
		},
		"with single parent": {
			currBranch: "feat/pt1",
			log: gitlib.Log{
				Commits: []gitlib.Commit{
					{
						Hash:          "c2",
						ParentHashes:  []string{"c1"},
						LocalBranches: []string{"feat/pt2"},
					},
					{
						Hash:          "c1",
						ParentHashes:  []string{"c0"},
						LocalBranches: []string{"feat/pt1"},
					},
				},
			},
			want: stackslib.Stacks{
				Entries: []stackslib.Stack{
					{
						Commits: []stackslib.Commit{
							{
								Hash:          "c2",
								LocalBranches: []string{"feat/pt2"},
							},
							{
								Hash:          "c1",
								LocalBranches: []string{"feat/pt1"},
							},
						},
						LocalBranches: []stackslib.Branch{
							{
								Name: "feat/pt2",
							},
							{
								Current: true,
								Name:    "feat/pt1",
							},
						},
					},
				},
			},
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
						ParentHashes:  []string{"c0"},
						LocalBranches: []string{"feat/pt2"},
					},
					{
						Hash:          "c1",
						ParentHashes:  []string{"c0"},
						LocalBranches: []string{"feat/pt1"},
					},
				},
			},
			wantErrContains: "multiple parents",
		},
		"multiple sources": {
			currBranch: "featA/pt2",
			log: gitlib.Log{
				Commits: []gitlib.Commit{
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
			want: stackslib.Stacks{
				Entries: []stackslib.Stack{
					{
						Commits: []stackslib.Commit{
							{
								Hash:          "c4",
								LocalBranches: []string{"featA/pt2"},
							},
							{
								Hash:          "c3",
								LocalBranches: nil,
							},
							{
								Hash:          "c2",
								LocalBranches: []string{"featA/pt1"},
							},
						},
						LocalBranches: []stackslib.Branch{
							{
								Current: true,
								Name:    "featA/pt2",
							},
							{
								Name: "featA/pt1",
							},
						},
					},
					{
						Commits: []stackslib.Commit{
							{
								Hash:          "c1",
								LocalBranches: []string{"featB/pt1"},
							},
						},
						LocalBranches: []stackslib.Branch{
							{
								Name: "featB/pt1",
							},
						},
					},
				},
			},
		},
		"multiple children": {
			currBranch: "feat/pt2",
			log: gitlib.Log{
				Commits: []gitlib.Commit{
					{
						Hash:          "c4",
						ParentHashes:  []string{"c3a"},
						LocalBranches: []string{"feat/pt3"},
					},
					{
						Hash:          "c3a",
						ParentHashes:  []string{"c2"},
						LocalBranches: nil,
					},
					{
						Hash:          "c3b",
						ParentHashes:  []string{"c2"},
						LocalBranches: []string{"feat/pt2"},
					},
					{
						Hash:          "c2",
						ParentHashes:  []string{"c0"},
						LocalBranches: []string{"feat/pt1"},
					},
				},
			},
			want: stackslib.Stacks{
				Entries: []stackslib.Stack{
					{
						Commits: []stackslib.Commit{
							{
								Hash:          "c3b",
								LocalBranches: []string{"feat/pt2"},
							},
							{
								Hash:          "c2",
								LocalBranches: []string{"feat/pt1"},
							},
						},
						LocalBranches: []stackslib.Branch{
							{
								Current: true,
								Name:    "feat/pt2",
							},
							{
								Name: "feat/pt1",
							},
						},
					},
					{
						Commits: []stackslib.Commit{
							{
								Hash:          "c4",
								LocalBranches: []string{"feat/pt3"},
							},
							{
								Hash: "c3a",
							},
							{
								Hash:          "c2",
								LocalBranches: []string{"feat/pt1"},
							},
						},
						LocalBranches: []stackslib.Branch{
							{
								Name: "feat/pt3",
							},
							{
								Name: "feat/pt1",
							},
						},
					},
				},
				SharingHistory: [][]string{
					{
						"feat/pt2",
						"feat/pt3",
					},
				},
			},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			fg := &FakeGit{
				Log:           c.log,
				CurrentBranch: c.currBranch,
			}

			got, err := stackslib.Compute(fg, "main")
			if c.wantErrContains != "" {
				require.ErrorContains(t, err, c.wantErrContains)
				return
			}
			require.NoError(t, err)
			require.Equal(t, c.want, got)
		})
	}
}

type FakeGit struct {
	Log           gitlib.Log
	CurrentBranch string
}

func (fg *FakeGit) LogAll(notReachableFrom string) (gitlib.Log, error) {
	return fg.Log, nil
}

func (fg *FakeGit) GetCurrentBranch() (string, error) {
	return fg.CurrentBranch, nil
}
