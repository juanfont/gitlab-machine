package executor

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/dimchansky/utfbom"
	"github.com/rs/zerolog/log"

	"github.com/juanfont/gitlab-machine/pkg/drivers"
)

type Executor struct {
	driver drivers.Driver
}

func NewExecutor(d drivers.Driver) (*Executor, error) {
	e := Executor{}
	e.driver = d
	return &e, nil
}

// Prepare calls the driver to ready up a new execution environment
func (e *Executor) Prepare() error {
	err := e.driver.Create()
	if err != nil {
		return err
	}

	log.Info().Msg("Setting up base software")
	if os, _ := e.driver.GetOS(); os == drivers.Windows {
		pw := `powershell New-ItemProperty -Path "HKLM:\SOFTWARE\OpenSSH" -Name DefaultShell -Value "C:\Windows\System32\WindowsPowerShell\v1.0\powershell.exe" -PropertyType String -Force`
		err = e.runCommand(pw, false)
		if err != nil {
			return err
		}

		err = e.runCommand("choco install -y --no-progress git.install;", false)
		if err != nil {
			return err
		}

		err = e.runCommand("refreshenv;", false)
		if err != nil {
			return err
		}

		err = e.runCommand("choco install -y --no-progress poshgit;", false)
		if err != nil {
			return err
		}

		err = e.runCommand("choco install -y --no-progress gitlab-runner;", false)
		if err != nil {
			return err
		}

		err = e.runCommand("Restart-Service -force sshd", false) // https://github.com/chocolatey/choco/issues/2694
		if err != nil {
			return err
		}
	}

	return nil
}

// Run executes the required script
func (e *Executor) Run(filePath string, stage string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	reader, _ := utfbom.Skip(bytes.NewReader(data))
	buf, err := io.ReadAll(reader)
	if err != nil {
		return err
	}

	log.Debug().Msgf("Starting stage on %s %s (%s)", e.driver.GetMachineName(), stage, filePath)
	err = e.runCommand(string(buf), true)
	if err != nil {
		return err
	}

	return nil
}

// Cleanup releases the resources once the job has finished
func (e *Executor) CleanUp() error {
	err := e.driver.Destroy()
	return err
}

// Shell opens a shell with the specified command
func (e *Executor) Shell(cmd string) error {
	client, err := e.driver.GetSSHClientFromDriver()
	if err != nil {
		return err
	}
	return client.Shell(cmd)
}

func (e *Executor) runCommand(command string, printOutput bool) error {
	client, err := e.driver.GetSSHClientFromDriver()
	if err != nil {
		return err
	}

	log.Debug().Str("command", command).Msg("Running command")

	output, err := client.Output(command)
	if err != nil {
		log.Error().
			Err(err).
			Str("command", command).
			Str("output", string(output)).
			Msg("Error running command")
		return fmt.Errorf("ssh command error")
	}

	if printOutput {
		fmt.Printf("%s", output)
		log.Debug().Msg("Command executed successfully")
	} else {
		log.Debug().Str("output", output).Msg("Command executed successfully")
	}

	return nil
}
