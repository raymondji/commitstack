package sampleusage_test

import (
	"os"
	"testing"

	"github.com/raymondji/commitstack/config"
	"github.com/raymondji/commitstack/githost"
	"github.com/raymondji/commitstack/libgit"
	"github.com/raymondji/commitstack/sampleusage"
	"github.com/stretchr/testify/require"
)

func TestBasics(t *testing.T) {
	if isCI() {
		t.Skip()
	}

	git, err := libgit.New()
	require.NoError(t, err)

	remote, err := git.GetRemote()
	require.NoError(t, err)

	cfg, err := config.Load()
	require.NoError(t, err)

	repoCfg, ok := cfg.Repositories[remote.URLPath]
	require.True(t, ok, "repo config is not setup for %s", remote.URLPath)

	host, err := githost.New(remote.Kind, repoCfg)
	require.NoError(t, err)

	basics := sampleusage.Basics(git, host, repoCfg.DefaultBranch, config.NewTheme(cfg.Theme))
	err = basics.Cleanup()
	require.NoError(t, err)
	err = basics.Execute()
	require.NoError(t, err)
}

func isCI() bool {
	return os.Getenv("GITHUB_ACTIONS") == "true"
}
