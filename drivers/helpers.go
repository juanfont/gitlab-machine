package drivers

import (
	"fmt"
	"time"

	"github.com/prometheus/common/log"
)

// From github.com/docker/machine/libmachine

func sshAvailableFunc(d Driver) func() bool {
	return func() bool {
		log.Debug("Getting to WaitForSSH function...")

		if _, err := runSSHCommandFromDriver(d, "exit"); err != nil {
			log.Debugf("Error getting ssh command 'exit' : %s", err)
			return false
		}
		return true
	}
}

func WaitForSSH(d Driver) error {
	// Try to dial SSH for 30 seconds before timing out.
	if err := waitForSpecific(sshAvailableFunc(d), 1200, 10*time.Second); err != nil {
		return fmt.Errorf("Too many retries waiting for SSH to be available. Last error: %s", err)
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

	log.Debugf("About to run SSH command:\n%s", command)

	output, err := client.Output(command)
	log.Debugf("SSH cmd err, output: %v: %s", err, output)
	if err != nil {
		return "", fmt.Errorf(`ssh command error:
command : %s
err     : %v
output  : %s`, command, err, output)
	}

	return output, nil
}
