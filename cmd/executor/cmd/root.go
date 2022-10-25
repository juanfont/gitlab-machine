package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	vcdcmd "github.com/juanfont/gitlab-machine/cmd/executor/cmd/vcd"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const version = "0.1"

var cfgFile string

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file path")

	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(vcdcmd.VcdCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath(".")
		viper.AddConfigPath("/opt/gitlab-machine")
		ex, err := os.Executable()
		if err != nil {
			panic(err)
		}
		exPath := filepath.Dir(ex)
		viper.AddConfigPath(exPath)
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		log.Fatal().Err(err).Msg("Error reading config file")
	}

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
		NoColor:    false,
	})

	logLevel := viper.GetString("log_level")
	if logLevel == "" {
		logLevel = "debug"
	}
	level, err := zerolog.ParseLevel(logLevel)
	if err != nil {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(level)
	}
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version.",
	Long:  "The version of the executor.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(version)
	},
}

var rootCmd = &cobra.Command{
	Use:   "executor",
	Short: "executor - a Gitlab Custom Executor",
	Long: `
A custom executor for Gitlab
Juan Font Alonso <juanfontalonso@gmail.com> - 2022
https://gitlab.com/juanfont/gitlab-machine`,
}
