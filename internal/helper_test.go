package internal

import (
	"io/fs"
	"testing"

	"github.com/stretchr/testify/require"
)


func fsEqual(t *testing.T, left fs.FS, right fs.FS) {
	t.Helper()

	require.NoError(t, fs.WalkDir(left, ".", func(path string, d fs.DirEntry, e1 error) error {
		if e1 != nil {
			return e1
		}

		if d.IsDir() {
			return nil
		}

		leftContent, err := fs.ReadFile(left, path)
		require.NoError(t, err)
		rightContent, err := fs.ReadFile(right, path)
		require.NoError(t, err)

		require.Equal(t, string(leftContent), string(rightContent))

		return nil
	}))
}
