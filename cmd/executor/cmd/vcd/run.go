package vcdcmd

import (
	"fmt"

	executor "github.com/juanfont/gitlab-machine"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var runVcdCmd = &cobra.Command{
	Use:   "run PATH STAGE",
	Short: "Run phase of the custom executor",
	Long:  "",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return fmt.Errorf("missing parameters")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		vcd := getVcdDriver()
		e, _ := executor.NewExecutor(vcd)
		err := e.Run(args[0], args[1])
		if err != nil {
			log.Fatal().Err(err).Msg("Error running the command")
		}
	},
}
