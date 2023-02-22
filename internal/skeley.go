package internal

import (
	"html/template"
	"os"
	"path/filepath"

	"github.com/rs/zerolog"
	"golang.org/x/mod/modfile"
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
	Logger           zerolog.Logger
	TemplateLocation string
}

func NewSkeley(conf SkeleyConfig) *Skeley {
	return &Skeley{
		log:  conf.Logger,
		conf: conf,
	}
}

type Skeley struct {
	log  zerolog.Logger
	conf SkeleyConfig
}

func (s *Skeley) ListTemplates() ([]string, error) {
	return nil, nil
}

func (s *Skeley) Execute(name string) error {
	templateDir, err := s.getTemplateDir()
	if err != nil {
		return err
	}

	selectedTemplate := filepath.Join(templateDir, name)
	config, err := s.getTemplateConfig(selectedTemplate)
	if err != nil {
		return err
	}

	vars := templateVars{}
	if !config.NotModule {
		mod, err := s.parseModule()
		if err != nil {
			return err
		}
		vars.Module = mod.Module
		vars.BinaryName = mod.BinaryName
		vars.GoVersion = mod.GoVersion
	}

	root, files, err := s.findAndParseTemplates(filepath.Join(selectedTemplate, "files"), template.FuncMap{})
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

func (s *Skeley) getTemplateDir() (string, error) {
	if s.conf.TemplateLocation != "" {
		return s.conf.TemplateLocation, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		s.log.Err(err).Msg("error getting user home dir")
		return "", err
	}
	return filepath.Join(home, ".config", "skeley", "templates"), nil
}

func (s *Skeley) getTemplateConfig(templatePath string) (templateConfig, error) {
	path := filepath.Join(templatePath, "config.yaml")
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			s.log.Debug().Str("path", path).Msg("no config found")
			return templateConfig{}, nil
		}
		s.log.Err(err).Msg("error checking for config file")
		return templateConfig{}, err
	}

	fBytes, err := os.ReadFile(path)
	if err != nil {
		s.log.Err(err).Msg("error reading config file")
		return templateConfig{}, err
	}

	var conf templateConfig
	if err := yaml.Unmarshal(fBytes, &conf); err != nil {
		s.log.Err(err).Msg("error unmarshalling")
		return templateConfig{}, err
	}

	return conf, nil
}

func (s *Skeley) parseModule() (moduleInfo, error) {
	modBytes, err := os.ReadFile("go.mod")
	if err != nil {
		s.log.Err(err).Msg("error reading `go.mod`")
		return moduleInfo{}, err
	}

	fl, err := modfile.Parse("go.mod", modBytes, nil)
	if err != nil {
		s.log.Err(err).Msg("error parsing `go.mod`")
		return moduleInfo{}, err
	}

	return moduleInfo{
		Module: fl.Module.Mod.Path,
		GoVersion: fl.Go.Version,
		BinaryName: filepath.Base(fl.Module.Mod.Path),
	}, nil
}

func (s *Skeley) findAndParseTemplates(rootDir string, funcMap template.FuncMap) (*template.Template, []string, error) {
	cleanRoot := filepath.Clean(rootDir)
	pfx := len(cleanRoot) + 1
	root := template.New("")

	filenames := []string{}

	err := filepath.Walk(cleanRoot, func(path string, info os.FileInfo, e1 error) error {
		if e1 != nil {
			return e1
		}
		if !info.IsDir() {
			b, e2 := os.ReadFile(path)
			if e2 != nil {
				s.log.Err(e2).Str("path", path).Msg("reading template file")
				return e2
			}

			name := path[pfx:]
			filenames = append(filenames, name)
			t := root.New(name).Funcs(funcMap)
			_, e2 = t.Parse(string(b))
			if e2 != nil {
				s.log.Err(e2).Str("path", path).Msg("parsing template file")
				return e2
			}
		}

		return nil
	})

	return root, filenames, err
}

func (s *Skeley) renderFile(tmpl *template.Template, name string, vars templateVars) error {
	if err := os.MkdirAll(filepath.Dir(name), 0775); err != nil {
		s.log.Err(err).Msg("making containing directory")
		return err
	}

	f, err := os.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0664)
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
