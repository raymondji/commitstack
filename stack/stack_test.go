package stack_test

import (
	"slices"
	"testing"

	"github.com/raymondji/git-stacked/gitlib"
	"github.com/raymondji/git-stacked/stack"
	"github.com/stretchr/testify/require"
)

func TestGetAll(t *testing.T) {
	cases := map[string]struct {
		currBranch      string
		log             gitlib.Log
		want            []stack.Stack
		wantErrContains string
	}{
		"empty": {
			log: gitlib.Log{
				Commits: nil,
			},
			currBranch: "main",
			want:       nil,
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
			want: []stack.Stack{
				{
					Name:    "dev",
					Current: true,
					LocalBranches: []stack.Branch{
						{
							Current: true,
							Name:    "dev",
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
			want: []stack.Stack{
				{
					Name:    "feat/pt2",
					Current: true,
					LocalBranches: []stack.Branch{
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
			want: []stack.Stack{
				{
					Name:    "featA/pt2",
					Current: true,
					LocalBranches: []stack.Branch{
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
					Name: "featB/pt1",
					LocalBranches: []stack.Branch{
						{
							Name: "featB/pt1",
						},
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

			got, err := stack.GetAll(fg, "main")
			if c.wantErrContains != "" {
				require.ErrorContains(t, err, c.wantErrContains)
				return
			}
			require.NoError(t, err)
			slices.SortFunc(got, func(a, b stack.Stack) int {
				if a.Name < b.Name {
					return -1
				} else if a.Name == b.Name {
					return 0
				} else {
					return 1
				}
			})
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
