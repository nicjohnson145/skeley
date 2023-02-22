package cmd

import (
	"fmt"

	"github.com/nicjohnson145/skeley/config"
	"github.com/nicjohnson145/skeley/internal"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func List() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "list",
		Short: "List available templates",
		Args: cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			skeley := internal.NewSkeley(internal.SkeleyConfig{
				Logger: config.InitLogger(),
				TemplateLocation: viper.GetString(config.TemplateDir),
			})
			tmpls, err := skeley.ListTemplates()
			if err != nil {
				return err
			}

			for _, t := range tmpls {
				fmt.Fprintln(cmd.OutOrStdout(), t)
			}

			return nil
		},
	}

	return rootCmd
}
