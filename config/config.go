package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

//go:generate go-enum -f $GOFILE -marshal -names -flag

/*
ENUM(
local
git
)
*/
type SourceType string

const (
	Debug           = "debug"
	TemplateDir     = "template-dir"
	OutputDirectory = "output-dir"
	InputType       = "input-type"
	KeyPath         = "key-path"
	Token           = "token"
	TokenUser       = "token-user"
)

const (
	DefaultDebug           = false
	DefaultOutputDirectory = "."
	DefaulInputType        = SourceTypeLocal
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
	viper.SetDefault(InputType, DefaulInputType)

	viper.BindPFlags(cmd.Flags())

	return nil
}
