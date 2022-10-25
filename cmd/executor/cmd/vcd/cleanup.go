package vcdcmd

import (
	executor "github.com/juanfont/gitlab-machine"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var cleanupVcdCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "Remove the current executor",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		vcdDriver, err := getVcdDriver()
		if err != nil {
			log.Fatal().Err(err).Msg("Error creating vcd driver")
		}
		e, _ := executor.NewExecutor(vcdDriver)
		err = e.CleanUp()
		if err != nil {
			log.Fatal().Err(err).Msg("Error cleaning up executor")
		}
	},
}
