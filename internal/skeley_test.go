package internal

import (
	"html/template"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/lithammer/dedent"
	"github.com/psanford/memfs"
	"github.com/stretchr/testify/require"
)

func TestListTemplates(t *testing.T) {
	t.Run("smokes", func(t *testing.T) {
		inp := memfs.New()
		require.NoError(t, inp.MkdirAll("template1", 0775))
		require.NoError(t, inp.MkdirAll("template2", 0775))
		require.NoError(t, inp.WriteFile("some-file", []byte("content"), 0664))

		sk := NewSkeley(SkeleyConfig{
			InputFS: inp,
		})

		templates, err := sk.ListTemplates()
		require.NoError(t, err)

		sort.Strings(templates)
		require.Equal(
			t,
			[]string{
				"template1",
				"template2",
			},
			templates,
		)
	})
}

func TestGetTemplateConfig(t *testing.T) {
	t.Run("with config", func(t *testing.T) {
		inpFS := memfs.New()
		require.NoError(t, inpFS.WriteFile(
			"config.yaml",
			[]byte(dedent.Dedent(`
				not-module: true
			`)),
			0644,
		))

		sk := NewSkeley(SkeleyConfig{
			InputFS: inpFS,
		})

		conf, err := sk.getTemplateConfig()
		require.NoError(t, err)

		require.Equal(
			t,
			templateConfig{
				NotModule: true,
			},
			conf,
		)
	})

	t.Run("no config", func(t *testing.T) {
		inpFS := memfs.New()

		sk := NewSkeley(SkeleyConfig{
			InputFS: inpFS,
		})

		conf, err := sk.getTemplateConfig()
		require.NoError(t, err)
		require.Equal(t, templateConfig{}, conf)
	})
}

func TestParseModule(t *testing.T) {
	testData := []struct {
		name     string
		mod      string
		expected moduleInfo
	}{
		{
			name: "smokes",
			mod: `
				module github.com/nicjohnson145/skeley

				go 1.20
			`,
			expected: moduleInfo{
				Module:     "github.com/nicjohnson145/skeley",
				GoVersion:  "1.20",
				BinaryName: "skeley",
			},
		},
		{
			name: "short name",
			mod: `
				module skeley

				go 1.20
			`,
			expected: moduleInfo{
				Module:     "skeley",
				GoVersion:  "1.20",
				BinaryName: "skeley",
			},
		},
		{
			name: "triple version",
			mod: `
				module skeley

				go 1.21.1
			`,
			expected: moduleInfo{
				Module:     "skeley",
				GoVersion:  "1.21.1",
				BinaryName: "skeley",
			},
		},
	}
	for _, tc := range testData {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			require.NoError(t, os.WriteFile(
				filepath.Join(dir, "go.mod"),
				[]byte(dedent.Dedent(tc.mod)),
				0644,
			))

			sk := NewSkeley(SkeleyConfig{
				OutputPath: dir,
			})

			info, err := sk.parseModule()
			require.NoError(t, err)
			require.Equal(t, tc.expected, info)
		})
	}
}

func TestFindAndParseTemplates(t *testing.T) {
	t.Run("smokes", func(t *testing.T) {
		sk := NewSkeley(SkeleyConfig{})

		inpFS := os.DirFS("./testdata/simple-module/input/files")

		_, files, err := sk.findAndParseTemplates(inpFS, template.FuncMap{})
		require.NoError(t, err)
		require.NotEmpty(t, files)
	})
}

func TestExecute(t *testing.T) {
	t.Run("simple module", func(t *testing.T) {
		dir := t.TempDir()
		destFS := os.DirFS(dir)
		expectedFS := os.DirFS("./testdata/simple-module/output")

		// Simulate running `go mod init` in the target repo
		modContent, err := fs.ReadFile(expectedFS, "go.mod")
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(filepath.Join(dir, "go.mod"), modContent, 0664))

		inpFS := os.DirFS("./testdata/simple-module/input")

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
