package gitmod_test

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wow-look-at-my/slopfix/gitmod"
)

// repoRoot answers this repository's own work tree.
func repoRoot(t *testing.T) string {
	t.Helper()
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		t.Skip("not a git work tree")
	}
	return strings.TrimSpace(string(out))
}

// Every submodule declares a rev, and it is the commit the index records.
//
// The two drift silently otherwise: a `git submodule update --remote` moves the
// gitlink and leaves .gitmodules alone, and a consumer resolving this module
// from the proxy then generates its tables from a DIFFERENT grammar than this
// repository's own CI does.
func TestPinnedRevMatchesTheGitlink(t *testing.T) {
	root := repoRoot(t)

	pins, err := gitmod.Pins(root)
	require.NoError(t, err)
	require.NotEmpty(t, pins, "this repository declares submodules")

	for _, pin := range pins {
		t.Run(pin.Path, func(t *testing.T) {
			assert.NotEmpty(t, pin.URL, "%s declares no url", pin.Path)
			require.NotEmpty(t, pin.Rev,
				"%s declares no rev. A module zip carries .gitmodules and no submodule, so the"+
					" rev is the only thing a consumer can fetch the grammar by", pin.Path)

			gitlink := gitmod.GitlinkRev(root, pin.Path)
			require.NotEmpty(t, gitlink, "%s has no gitlink in the index", pin.Path)
			assert.Equal(t, gitlink, pin.Rev,
				"%s: .gitmodules pins %s and the index records %s. Update the rev key to match",
				pin.Path, pin.Rev, gitlink)
		})
	}
}
