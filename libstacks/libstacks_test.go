package libstacks_test

import (
	"testing"

	"github.com/raymondji/git-stack/libgit"
	"github.com/raymondji/git-stack/libstacks"
	"github.com/stretchr/testify/require"
)

func TestCompute(t *testing.T) {
	cases := map[string]struct {
		currBranch      string
		log             libgit.Log
		want            libstacks.Stacks
		wantErrContains string
	}{
		"empty": {
			log: libgit.Log{
				Commits: nil,
			},
			currBranch: "main",
			want:       libstacks.Stacks{},
		},
		"one": {
			currBranch: "dev",
			log: libgit.Log{
				Commits: []libgit.Commit{
					{
						Hash:          "c1",
						ParentHashes:  []string{"p1"},
						LocalBranches: []string{"dev"},
					},
				},
			},
			want: libstacks.Stacks{
				Entries: []libstacks.Stack{
					{
						Commits: []libstacks.Commit{
							{
								Hash: "c1",
								LocalBranch: &libstacks.Branch{
									Name:    "dev",
									Current: true,
								},
							},
						},
					},
				},
			},
		},
		"with single parent": {
			currBranch: "feat/pt1",
			log: libgit.Log{
				Commits: []libgit.Commit{
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
			want: libstacks.Stacks{
				Entries: []libstacks.Stack{
					{
						Commits: []libstacks.Commit{
							{
								Hash: "c2",
								LocalBranch: &libstacks.Branch{
									Name: "feat/pt2",
								},
							},
							{
								Hash: "c1",
								LocalBranch: &libstacks.Branch{
									Name:    "feat/pt1",
									Current: true,
								},
							},
						},
					},
				},
			},
		},
		"with multiple parents": {
			currBranch: "feat/pt2",
			log: libgit.Log{
				Commits: []libgit.Commit{
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
			want: libstacks.Stacks{
				Errors: []error{
					libstacks.MergeCommitError{
						MergeCommitHash: "c3",
						PartialStack: libstacks.Stack{
							Commits: []libstacks.Commit{
								{
									Hash: "c3",
									LocalBranch: &libstacks.Branch{
										Name: "feat/pt3",
									},
								},
								{
									Hash: "c2",
									LocalBranch: &libstacks.Branch{
										Name:    "feat/pt2",
										Current: true,
									},
								},
							},
						},
					},
					libstacks.MergeCommitError{
						MergeCommitHash: "c3",
						PartialStack: libstacks.Stack{
							Commits: []libstacks.Commit{
								{
									Hash: "c3",
									LocalBranch: &libstacks.Branch{
										Name: "feat/pt3",
									},
								},
								{
									Hash: "c1",
									LocalBranch: &libstacks.Branch{
										Name: "feat/pt1",
									},
								},
							},
						},
					},
				},
			},
		},
		"multiple sources": {
			currBranch: "featA/pt2",
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
			want: libstacks.Stacks{
				Entries: []libstacks.Stack{
					{
						Commits: []libstacks.Commit{
							{
								Hash: "c4",
								LocalBranch: &libstacks.Branch{
									Name:    "featA/pt2",
									Current: true,
								},
							},
							{
								Hash: "c3",
							},
							{
								Hash: "c2",
								LocalBranch: &libstacks.Branch{
									Name: "featA/pt1",
								},
							},
						},
					},
					{
						Commits: []libstacks.Commit{
							{
								Hash: "c1",
								LocalBranch: &libstacks.Branch{
									Name: "featB/pt1",
								},
							},
						},
					},
				},
			},
		},
		"multiple children": {
			currBranch: "feat/pt2",
			log: libgit.Log{
				Commits: []libgit.Commit{
					{
						Hash:          "c4",
						ParentHashes:  []string{"c3a"},
						LocalBranches: []string{"feat/pt3"},
					},
					{
						Hash:         "c3a",
						ParentHashes: []string{"c2"},
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
			want: libstacks.Stacks{
				Entries: []libstacks.Stack{
					{
						Commits: []libstacks.Commit{
							{
								Hash: "c3b",
								LocalBranch: &libstacks.Branch{
									Name:    "feat/pt2",
									Current: true,
								},
							},
							{
								Hash: "c2",
								LocalBranch: &libstacks.Branch{
									Name: "feat/pt1",
								},
							},
						},
						Error: libstacks.SharedCommitError{
							StackNames: []string{"feat/pt2", "feat/pt3"},
						},
					},
					{
						Commits: []libstacks.Commit{
							{
								Hash: "c4",
								LocalBranch: &libstacks.Branch{
									Name: "feat/pt3",
								},
							},
							{
								Hash: "c3a",
							},
							{
								Hash: "c2",
								LocalBranch: &libstacks.Branch{
									Name: "feat/pt1",
								},
							},
						},
						Error: libstacks.SharedCommitError{
							StackNames: []string{"feat/pt2", "feat/pt3"},
						},
					},
				},
				Errors: []error{
					libstacks.SharedCommitError{
						StackNames: []string{"feat/pt2", "feat/pt3"},
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

			got, err := libstacks.Compute(fg, "main")
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
	Log           libgit.Log
	CurrentBranch string
}

func (fg *FakeGit) LogAll(notReachableFrom string) (libgit.Log, error) {
	return fg.Log, nil
}

func (fg *FakeGit) GetCurrentBranch() (string, error) {
	return fg.CurrentBranch, nil
}
