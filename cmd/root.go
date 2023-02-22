package cmd

import (
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
			skeley := internal.NewSkeley(internal.SkeleyConfig{
				Logger: config.InitLogger(),
				TemplateLocation: viper.GetString(config.TemplateDir),
			})
			return skeley.Execute(args[0])
		},
	}
	rootCmd.PersistentFlags().BoolP(config.Debug, "d", false, "Enable debug logging")
	rootCmd.PersistentFlags().StringP(config.TemplateDir, "t", "", "Override default template directory")

	rootCmd.AddCommand(
		List(),
	)

	return rootCmd
}
