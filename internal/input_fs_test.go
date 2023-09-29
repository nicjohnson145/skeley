package internal

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/nicjohnson145/skeley/config"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestGitFSIntegration(t *testing.T) {
	if _, ok := os.LookupEnv("INTEGRATION_TEST"); !ok {
		t.Skip("skipping due to INTEGRATION_TEST not set")
	}

	
	t.Run("simple module", func(t *testing.T) {
		dir := t.TempDir()
		destFS := os.DirFS(dir)
		expectedFS := os.DirFS("./testdata/simple-module/output")

		// Simulate running `go mod init` in the target repo
		modContent, err := fs.ReadFile(expectedFS, "go.mod")
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(filepath.Join(dir, "go.mod"), modContent, 0664))

		viper.Set(config.TemplateDir, "https://github.com/nicjohnson145/skeley-remote-template-example.git")
		viper.Set(config.InputType, config.SourceTypeGit)

		inpFS, err := InputFSFromEnv(zerolog.Logger{}, "example")
		require.NoError(t, err)

		sk := NewSkeley(SkeleyConfig{
			InputFS:    inpFS,
			OutputPath: dir,
		})

		require.NoError(t, sk.Execute())

		// Everything in the output directory matches the expected (no extra files)
		fsEqual(t, destFS, expectedFS)
		// Everything in the expected directory matches the actual (no missing files)
		fsEqual(t, expectedFS, destFS)
	})
}
