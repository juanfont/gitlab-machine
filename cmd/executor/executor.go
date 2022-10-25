package main

import (
	"os"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/juanfont/gitlab-machine/cmd/executor/cmd"
	"github.com/rs/zerolog"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
		NoColor:    false,
	})

	cmd.Execute()
}
