package internal

import (
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/rs/zerolog"
	"golang.org/x/mod/modfile"
	"golang.org/x/mod/semver"
	"gopkg.in/yaml.v3"
)

type templateConfig struct {
	PreCmds   []string `yaml:"pre-cmds,omitempty"`
	PostCmds  []string `yaml:"post-cmds,omitempty"`
	NotModule bool     `yaml:"not-module"`
}

type moduleInfo struct {
	Module     string
	BinaryName string
	GoVersion  string
}

type templateVars struct {
	Module     string
	BinaryName string
	GoVersion  string
}

type SkeleyConfig struct {
	Logger     zerolog.Logger
	InputFS    fs.FS
	OutputPath string
}

func NewSkeley(conf SkeleyConfig) *Skeley {
	return &Skeley{
		log:        conf.Logger,
		conf:       conf,
		inputFS:    conf.InputFS,
		outputPath: conf.OutputPath,
	}
}

type Skeley struct {
	log        zerolog.Logger
	conf       SkeleyConfig
	inputFS    fs.FS
	outputPath string
}

func (s *Skeley) ListTemplates() ([]string, error) {
	entries, err := fs.ReadDir(s.inputFS, ".")
	if err != nil {
		return nil, fmt.Errorf("error listing directory: %w", err)
	}

	templates := []string{}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}

		templates = append(templates, e.Name())
	}

	return templates, nil
}

func (s *Skeley) Execute() error {
	s.log.Debug().Msg("attempting to read template config")
	config, err := s.getTemplateConfig()
	if err != nil {
		return err
	}

	vars := templateVars{}
	if !config.NotModule {
		s.log.Debug().Msg("template is configured as go module, attempting to parse go.mod")
		mod, err := s.parseModule()
		if err != nil {
			return err
		}
		vars.Module = mod.Module
		vars.BinaryName = mod.BinaryName
		vars.GoVersion = mod.GoVersion
	}

	filesFS, err := fs.Sub(s.inputFS, "files")
	if err != nil {
		return fmt.Errorf("error creating subFS: %w", err)
	}

	root, files, err := s.findAndParseTemplates(filesFS, template.FuncMap{})
	if err != nil {
		return err
	}

	for _, fl := range files {
		if err := s.renderFile(root, fl, vars); err != nil {
			return err
		}
	}

	return nil
}

func (s *Skeley) getTemplateConfig() (templateConfig, error) {
	content, err := fs.ReadFile(s.inputFS, "config.yaml")
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return templateConfig{}, nil
		}
		return templateConfig{}, fmt.Errorf("error reading config: %w", err)
	}

	var conf templateConfig
	if err := yaml.Unmarshal(content, &conf); err != nil {
		s.log.Err(err).Msg("error unmarshalling")
		return templateConfig{}, err
	}

	return conf, nil
}

func (s *Skeley) parseModule() (moduleInfo, error) {
	modBytes, err := os.ReadFile(s.getGoModPath())
	if err != nil {
		s.log.Err(err).Msg("error reading `go.mod`")
		return moduleInfo{}, err
	}

	fl, err := modfile.Parse("go.mod", modBytes, func(path, version string) (string, error) {
		return semver.Canonical(version), nil
	})
	if err != nil {
		s.log.Err(err).Msg("error parsing `go.mod`")
		return moduleInfo{}, err
	}

	return moduleInfo{
		Module:     fl.Module.Mod.Path,
		GoVersion:  fl.Go.Version,
		BinaryName: filepath.Base(fl.Module.Mod.Path),
	}, nil
}

func (s *Skeley) getGoModPath() string {
	return filepath.Join(s.outputPath, "go.mod")
}

func (s *Skeley) findAndParseTemplates(fsys fs.FS, funcMap template.FuncMap) (*template.Template, []string, error) {
	root := template.New("")

	filenames := []string{}

	err := fs.WalkDir(fsys, ".", func(path string, info fs.DirEntry, e1 error) error {
		if e1 != nil {
			return fmt.Errorf("error from walk function: %w", e1)
		}
		if !info.IsDir() {
			b, e2 := fs.ReadFile(fsys, path)
			if e2 != nil {
				s.log.Err(e2).Str("path", path).Msg("reading template file")
				return e2
			}

			filenames = append(filenames, path)
			t := root.New(path).Funcs(funcMap)
			_, e2 = t.Parse(string(b))
			if e2 != nil {
				s.log.Err(e2).Str("path", path).Msg("parsing template file")
				return e2
			}
		}

		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	return root, filenames, nil
}

func (s *Skeley) renderFile(tmpl *template.Template, name string, vars templateVars) error {
	output := filepath.Join(s.outputPath, name)

	if err := os.MkdirAll(filepath.Dir(output), 0775); err != nil {
		s.log.Err(err).Msg("making containing directory")
		return err
	}

	f, err := os.OpenFile(output, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0664)
	if err != nil {
		s.log.Err(err).Msg("opening target file")
		return err
	}
	defer f.Close()

	if err := tmpl.ExecuteTemplate(f, name, vars); err != nil {
		s.log.Err(err).Msg("executing template")
		return err
	}

	return nil
}
