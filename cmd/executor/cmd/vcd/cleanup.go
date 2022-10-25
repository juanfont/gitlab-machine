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
		vcd := getVcdDriver()
		e, _ := executor.NewExecutor(vcd)
		err := e.CleanUp()
		if err != nil {
			log.Fatal().Err(err).Msg("Error cleaning up executor")
		}
	},
}
