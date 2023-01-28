package main

import (
	"html/template"
	"os"
	"strings"

	"path/filepath"
	"fmt"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"github.com/rs/zerolog/log"
)

type TemplateConfig struct {
	PreCmds  []string `yaml:"pre_cmds,omitempty"`
	Files    []string `yaml:"files,omitempty"`
	PostCmds []string `yaml:"post_cmds,omitempty"`
}

type TemplateVars struct {
	BinaryName string
	Module     string
}

func root() *cobra.Command {
	var templateDir string
	var debug bool

	root := &cobra.Command{
		Use: "skeley [OPTS] <template-name>",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			initializeConfig(cmd)
			cmd.SilenceUsage = true
		},
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			initLogging(debug)
			if templateDir == "" {
				home, err := os.UserHomeDir()
				if err != nil {
					log.Err(err).Msg("getting home directory")
					return fmt.Errorf("error getting user home directory: %w", err)
				}
				templateDir = filepath.Join(home, ".config", "skeley", "templates")
			}

			selectedTemplate := filepath.Join(templateDir, args[0])

			confBytes, err := os.ReadFile(filepath.Join(selectedTemplate, "config.yaml"))
			if err != nil {
				log.Err(err).Msg("reading skeley config")
				return err
			}

			var conf TemplateConfig
			if err := yaml.Unmarshal(confBytes, &conf); err != nil {
				log.Err(err).Msg("unmarshalling skeley config")
				return err
			}

			module, err := getModule()
			if err != nil {
				log.Err(err).Msg("reading go module information")
				return err
			}

			root, err := findAndParseTemplates(filepath.Join(templateDir, "files"), template.FuncMap{})
			if err != nil {
				log.Err(err).Msg("parsing template files")
				return err
			}

			tmplVars := TemplateVars{
				Module: module,
				BinaryName: filepath.Base(module),
			}
			for _, fl := range conf.Files {
				if err := renderFile(root, fl, tmplVars); err != nil {
					log.Err(err).Msg("rendering template files")
					return err
				}
			}

			return nil
		},
	}

	root.Flags().StringVarP(&templateDir, "template-dir", "t", "", "Override the default template directory")
	root.Flags().BoolVarP(&debug, "debug", "d", false, "Enable debug logging")

	return root
}

func renderFile(tmpl *template.Template, name string, vars TemplateVars) error {
	if err := os.MkdirAll(filepath.Dir(name), 0775); err != nil {
		log.Err(err).Msg("making containing directory")
		return err
	}

	f, err := os.OpenFile(name, os.O_RDWR | os.O_CREATE | os.O_TRUNC, 0664)
	if err != nil {
		log.Err(err).Msg("opening target file")
		return err
	}
	defer f.Close()

	if err := tmpl.ExecuteTemplate(f, name, vars); err != nil {
		log.Err(err).Msg("executing template")
		return err
	}

	return nil
}

func getModule() (string, error) {
	modBytes, err := os.ReadFile("go.mod")
	if err != nil {
		log.Err(err).Msg("reading go.mod")
		return "", err
	}

	modLines := strings.Split(string(modBytes), "\n")
	return strings.TrimPrefix(modLines[0], "module "), nil
}

func findAndParseTemplates(rootDir string, funcMap template.FuncMap) (*template.Template, error) {
	cleanRoot := filepath.Clean(rootDir)
	pfx := len(cleanRoot) + 1
	root := template.New("")

	err := filepath.Walk(cleanRoot, func(path string, info os.FileInfo, e1 error) error {
		if !info.IsDir() {
			if e1 != nil {
				return e1
			}

			b, e2 := os.ReadFile(path)
			if e2 != nil {
				log.Err(e2).Str("path", path).Msg("reading template file")
				return e2
			}

			name := path[pfx:]
			t := root.New(name).Funcs(funcMap)
			_, e2 = t.Parse(string(b))
			if e2 != nil {
				log.Err(e2).Str("path", path).Msg("parsing template file")
				return e2
			}
		}

		return nil
	})

	return root, err
}

func main() {
	if err := root().Execute(); err != nil {
		os.Exit(1)
	}
}
