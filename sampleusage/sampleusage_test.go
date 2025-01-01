package sampleusage_test

import (
	"testing"

	"github.com/raymondji/commitstack/config"
	"github.com/raymondji/commitstack/githost"
	"github.com/raymondji/commitstack/libgit"
	"github.com/raymondji/commitstack/sampleusage"
	"github.com/stretchr/testify/require"
)

func TestSamples(t *testing.T) {
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

	samples := sampleusage.New(config.NewTheme(cfg.Theme), repoCfg.DefaultBranch, git, host)
	err = samples.Cleanup()
	require.NoError(t, err)
	err = samples.Part1().Execute()
	require.NoError(t, err)
}
