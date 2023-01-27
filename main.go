package main

import (
	"html/template"
	"os"
	"strings"

	"path/filepath"
	"fmt"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

const (
	TemplateDir = "/home/njohnson/scratch/skeley-templates"
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
	return &cobra.Command{
		Use: "skeley",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			cmd.SilenceUsage = true
		},
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			templateDir := filepath.Join(TemplateDir, args[0])

			confBytes, err := os.ReadFile(filepath.Join(templateDir, "config.yaml"))
			if err != nil {
				fmt.Println("err reading config")
				return err
			}

			var conf TemplateConfig
			if err := yaml.Unmarshal(confBytes, &conf); err != nil {
				fmt.Println("err unmarshalling")
				return err
			}

			module, err := getModule()
			if err != nil {
				fmt.Println("err getting module")
				return err
			}

			root, err := findAndParseTemplates(filepath.Join(templateDir, "files"), template.FuncMap{})
			if err != nil {
				fmt.Println("err parsing templates")
				return err
			}

			tmplVars := TemplateVars{
				Module: module,
				BinaryName: filepath.Base(module),
			}
			for _, fl := range conf.Files {
				if err := renderFile(root, fl, tmplVars); err != nil {
					fmt.Println("err rendering templates")
					return err
				}
			}

			return nil
		},
	}
}

func renderFile(tmpl *template.Template, name string, vars TemplateVars) error {
	if err := os.MkdirAll(filepath.Dir(name), 0775); err != nil {
		return err
	}

	f, err := os.OpenFile(name, os.O_RDWR | os.O_CREATE | os.O_TRUNC, 0664)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := tmpl.ExecuteTemplate(f, name, vars); err != nil {
		return err
	}

	return nil
}

func getModule() (string, error) {
	modBytes, err := os.ReadFile("go.mod")
	if err != nil {
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
				return e2
			}

			name := path[pfx:]
			t := root.New(name).Funcs(funcMap)
			_, e2 = t.Parse(string(b))
			if e2 != nil {
				fmt.Printf("err parsing template %v\n", path)
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
