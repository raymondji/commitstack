package commitstack_test

import (
	"testing"

	"github.com/raymondji/git-stack/commitstack"
	"github.com/raymondji/git-stack/libgit"
	"github.com/stretchr/testify/require"
)

func TestCompute(t *testing.T) {
	cases := map[string]struct {
		currBranch      string
		log             libgit.Log
		want            commitstack.Stacks
		wantErrContains string
	}{
		"empty": {
			log: libgit.Log{
				Commits: nil,
			},
			currBranch: "main",
			want:       commitstack.Stacks{},
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
			want: commitstack.Stacks{
				Entries: []commitstack.Stack{
					{
						Commits: []commitstack.Commit{
							{
								Hash: "c1",
								LocalBranch: &commitstack.Branch{
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
			want: commitstack.Stacks{
				Entries: []commitstack.Stack{
					{
						Commits: []commitstack.Commit{
							{
								Hash: "c2",
								LocalBranch: &commitstack.Branch{
									Name: "feat/pt2",
								},
							},
							{
								Hash: "c1",
								LocalBranch: &commitstack.Branch{
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
			want: commitstack.Stacks{
				Errors: []error{
					commitstack.MergeCommitError{
						MergeCommitHash: "c3",
						PartialStack: commitstack.Stack{
							Commits: []commitstack.Commit{
								{
									Hash: "c3",
									LocalBranch: &commitstack.Branch{
										Name: "feat/pt3",
									},
								},
								{
									Hash: "c2",
									LocalBranch: &commitstack.Branch{
										Name:    "feat/pt2",
										Current: true,
									},
								},
							},
						},
					},
					commitstack.MergeCommitError{
						MergeCommitHash: "c3",
						PartialStack: commitstack.Stack{
							Commits: []commitstack.Commit{
								{
									Hash: "c3",
									LocalBranch: &commitstack.Branch{
										Name: "feat/pt3",
									},
								},
								{
									Hash: "c1",
									LocalBranch: &commitstack.Branch{
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
			want: commitstack.Stacks{
				Entries: []commitstack.Stack{
					{
						Commits: []commitstack.Commit{
							{
								Hash: "c4",
								LocalBranch: &commitstack.Branch{
									Name:    "featA/pt2",
									Current: true,
								},
							},
							{
								Hash: "c3",
							},
							{
								Hash: "c2",
								LocalBranch: &commitstack.Branch{
									Name: "featA/pt1",
								},
							},
						},
					},
					{
						Commits: []commitstack.Commit{
							{
								Hash: "c1",
								LocalBranch: &commitstack.Branch{
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
			want: commitstack.Stacks{
				Entries: []commitstack.Stack{
					{
						Commits: []commitstack.Commit{
							{
								Hash: "c3b",
								LocalBranch: &commitstack.Branch{
									Name:    "feat/pt2",
									Current: true,
								},
							},
							{
								Hash: "c2",
								LocalBranch: &commitstack.Branch{
									Name: "feat/pt1",
								},
							},
						},
						Error: commitstack.SharedCommitError{
							StackNames: []string{"feat/pt2", "feat/pt3"},
						},
					},
					{
						Commits: []commitstack.Commit{
							{
								Hash: "c4",
								LocalBranch: &commitstack.Branch{
									Name: "feat/pt3",
								},
							},
							{
								Hash: "c3a",
							},
							{
								Hash: "c2",
								LocalBranch: &commitstack.Branch{
									Name: "feat/pt1",
								},
							},
						},
						Error: commitstack.SharedCommitError{
							StackNames: []string{"feat/pt2", "feat/pt3"},
						},
					},
				},
				Errors: []error{
					commitstack.SharedCommitError{
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

			got, err := commitstack.ComputeAll(fg, "main")
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
