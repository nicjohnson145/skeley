package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func initLogging(debug bool) {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}
}

func initializeConfig(cmd *cobra.Command) {
	v := viper.New()
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	v.AutomaticEnv()
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		configName := f.Name
		if !f.Changed && v.IsSet(configName) {
			cmd.Flags().Set(f.Name, fmt.Sprint(v.Get(configName)))
		}
	})
}
