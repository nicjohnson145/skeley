package cmd

import (
	"fmt"
	"io/fs"
	"os"

	"github.com/nicjohnson145/skeley/config"
	"github.com/nicjohnson145/skeley/internal"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func Root() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "skeley [OPTS] <TEMPLATE_NAME>",
		Short: "Execute directory templates",
		Args: cobra.ExactArgs(1),
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// So we don't print usage messages on execution errors
			cmd.SilenceUsage = true
			// So we dont double report errors
			cmd.SilenceErrors = true
			return config.InitializeConfig(cmd)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			templates := os.DirFS(viper.GetString(config.TemplateDir))
			selected, err := fs.Sub(templates, args[0])
			if err != nil {
				return fmt.Errorf("error creating subFS: %w", err)
			}

			skeley := internal.NewSkeley(internal.SkeleyConfig{
				Logger: config.InitLogger(),
				InputFS: selected,
				OutputPath: viper.GetString(config.OutputDirectory),
			})
			return skeley.Execute()
		},
	}
	rootCmd.PersistentFlags().BoolP(config.Debug, "d", config.DefaultDebug, "Enable debug logging")
	rootCmd.PersistentFlags().StringP(config.TemplateDir, "t", "", "Override default template directory of '~/.config/skeley/templates'")

	rootCmd.Flags().StringP(config.OutputDirectory, "o", config.DefaultOutputDirectory, "Where to output the rendered template")

	rootCmd.AddCommand(
		List(),
	)

	return rootCmd
}
