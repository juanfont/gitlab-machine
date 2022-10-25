package vcdcmd

import (
	executor "github.com/juanfont/gitlab-machine"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var prepareVcdCmd = &cobra.Command{
	Use:   "prepare",
	Short: "Prepare a new instance of the vCloud Director executor",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		vcdDriver, err := getVcdDriver()
		if err != nil {
			log.Fatal().Err(err).Msg("Error creating vcd driver")
		}
		e, _ := executor.NewExecutor(vcdDriver)

		err = e.Prepare()
		if err != nil {
			log.Fatal().Err(err).Msg("Error preparing executor")
		}
	},
}
