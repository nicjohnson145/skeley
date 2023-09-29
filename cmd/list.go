package cmd

import (
	"fmt"

	"github.com/nicjohnson145/skeley/config"
	"github.com/nicjohnson145/skeley/internal"
	"github.com/spf13/cobra"
)

func List() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "list",
		Short: "List available templates",
		Args: cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			log := config.InitLogger()

			inputFS, err := internal.InputFSFromEnv(log, args[0])
			if err != nil {
				return err
			}

			skeley := internal.NewSkeley(internal.SkeleyConfig{
				Logger: log,
				InputFS: inputFS,
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
