package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	Debug           = "debug"
	TemplateDir     = "template-dir"
	OutputDirectory = "output-dir"
)

const (
	DefaultDebug           = false
	DefaultOutputDirectory = "."
)

func InitializeConfig(cmd *cobra.Command) error {
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("error getting user homedir: %w", err)
	}

	viper.SetDefault(TemplateDir, filepath.Join(home, ".config", "skeley", "templates"))

	viper.SetDefault(Debug, DefaultDebug)
	viper.SetDefault(OutputDirectory, DefaultOutputDirectory)

	viper.BindPFlags(cmd.Flags())

	return nil
}
