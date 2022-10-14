package drivers

import (
	"fmt"
	"time"

	"github.com/juanfont/gitlab-machine/pkg/utils"
	"github.com/rs/zerolog/log"
)

const (
	ErrExecutingSSHCommand  = utils.Error("error executing SSH command")
	ErrTooManyRetriesForSSH = utils.Error("too many retries waiting for SSH to be available")
)

// From github.com/docker/machine/libmachine

func sshAvailableFunc(d Driver) func() bool {
	return func() bool {
		log.Debug().Msg("Getting to WaitForSSH function...")

		if _, err := runSSHCommandFromDriver(d, "exit"); err != nil {
			log.Debug().Err(err).Msgf("Error getting ssh command 'exit'")
			return false
		}
		return true
	}
}

func WaitForSSH(d Driver) error {
	// Try to dial SSH for 30 seconds before timing out.
	if err := waitForSpecific(sshAvailableFunc(d), 1200, 10*time.Second); err != nil {
		log.Error().Err(err).Msg("Error waiting for SSH")
		return ErrTooManyRetriesForSSH
	}
	return nil
}

func waitForSpecific(f func() bool, maxAttempts int, waitInterval time.Duration) error {
	return waitForSpecificOrError(func() (bool, error) {
		return f(), nil
	}, maxAttempts, waitInterval)
}

func waitForSpecificOrError(f func() (bool, error), maxAttempts int, waitInterval time.Duration) error {
	for i := 0; i < maxAttempts; i++ {
		stop, err := f()
		if err != nil {
			return err
		}
		if stop {
			return nil
		}
		time.Sleep(waitInterval)
	}
	return fmt.Errorf("Maximum number of retries (%d) exceeded", maxAttempts)
}

func runSSHCommandFromDriver(d Driver, command string) (string, error) {
	client, err := d.GetSSHClientFromDriver()
	if err != nil {
		return "", err
	}

	log.Info().Msgf("Running SSH command: %s", command)

	output, err := client.Output(command)
	if err != nil {
		log.Error().
			Err(err).
			Str("output", output).
			Msgf("Error running SSH command")
		return "", ErrExecutingSSHCommand
	}

	return output, nil
}
