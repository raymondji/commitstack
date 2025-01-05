package commitstack_test

import (
	"fmt"
	"testing"

	"github.com/raymondji/git-stack-cli/commitstack"
	"github.com/raymondji/git-stack-cli/libgit"
	"github.com/stretchr/testify/require"
)

func TestInferStacks(t *testing.T) {
	cases := map[string]struct {
		log                            libgit.Log
		commitHashToContainingBranches map[string][]string
		want                           commitstack.InferenceResult
	}{
		"empty": {
			log: libgit.Log{
				Commits: nil,
			},
			want: commitstack.InferenceResult{},
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
			want: commitstack.InferenceResult{
				InferredStacks: []commitstack.Stack{
					{
						Commits: []commitstack.Commit{
							{
								Hash:          "c1",
								LocalBranches: []string{"dev"},
							},
						},
					},
				},
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
			want: commitstack.InferenceResult{
				InferredStacks: []commitstack.Stack{
					{
						Commits: []commitstack.Commit{
							{
								Hash:          "c1",
								LocalBranches: []string{"dev2", "dev"},
							},
						},
					},
				},
			},
		},
		"with single parent": {
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
			want: commitstack.InferenceResult{
				InferredStacks: []commitstack.Stack{
					{
						Commits: []commitstack.Commit{
							{
								Hash:          "c2",
								LocalBranches: []string{"feat/pt2"},
							},
							{
								Hash:          "c1",
								LocalBranches: []string{"feat/pt1"},
							},
						},
					},
				},
			},
		},
		"with multiple parents": {
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
			commitHashToContainingBranches: map[string][]string{
				"c3": {"feat/pt3"},
			},
			want: commitstack.InferenceResult{
				InferenceErrors: []error{
					commitstack.MergeCommitError{
						MergeCommitHash:    "c3",
						ContainingBranches: []string{"feat/pt3"},
					},
					commitstack.MergeCommitError{
						MergeCommitHash:    "c3",
						ContainingBranches: []string{"feat/pt3"},
					},
				},
			},
		},
		"multiple sources": {
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
			want: commitstack.InferenceResult{
				InferredStacks: []commitstack.Stack{
					{
						Commits: []commitstack.Commit{
							{
								Hash:          "c4",
								LocalBranches: []string{"featA/pt2"},
							},
							{
								Hash: "c3",
							},
							{
								Hash:          "c2",
								LocalBranches: []string{"featA/pt1"},
							},
						},
					},
					{
						Commits: []commitstack.Commit{
							{
								Hash:          "c1",
								LocalBranches: []string{"featB/pt1"},
							},
						},
					},
				},
			},
		},
		"multiple children": {
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
			want: commitstack.InferenceResult{
				InferredStacks: []commitstack.Stack{
					{
						Commits: []commitstack.Commit{
							{
								Hash:          "c3b",
								LocalBranches: []string{"feat/pt2"},
							},
							{
								Hash:          "c2",
								LocalBranches: []string{"feat/pt1"},
							},
						},
						ValidationErrors: []error{
							commitstack.DivergenceError{
								StackName:       "feat/pt2",
								OtherStackNames: []string{"feat/pt3"},
							},
						},
					},
					{
						Commits: []commitstack.Commit{
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
						ValidationErrors: []error{
							commitstack.DivergenceError{
								StackName:       "feat/pt3",
								OtherStackNames: []string{"feat/pt2"},
							},
						},
					},
				},
			},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			fg := &FakeGit{
				CommitHashToContainingBranches: c.commitHashToContainingBranches,
			}
			got, err := commitstack.InferStacks(fg, c.log)
			require.NoError(t, err)
			require.Equal(t, c.want, got)
		})
	}
}

type FakeGit struct {
	CommitHashToContainingBranches map[string][]string
}

func (fg *FakeGit) GetBranchesContainingCommit(commit string) ([]string, error) {
	if got, ok := fg.CommitHashToContainingBranches[commit]; !ok {
		return nil, fmt.Errorf("no entry found for commit %s", commit)
	} else {
		return got, nil
	}
}
