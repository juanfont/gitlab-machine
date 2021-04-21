package executor

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/dimchansky/utfbom"
	"github.com/juanfont/gitlab-machine/drivers"
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

	log.Println("Setting up base software")
	if os, _ := e.driver.GetOS(); os == drivers.Windows {
		pw := `powershell New-ItemProperty -Path "HKLM:\SOFTWARE\OpenSSH" -Name DefaultShell -Value "C:\Windows\System32\WindowsPowerShell\v1.0\powershell.exe" -PropertyType String -Force`
		e.runCommand(pw, false)
		e.runCommand("choco install -y --no-progress git git-lfs gitlab-runner", false)
	}
	return nil
}

// Run executes the required script
func (e *Executor) Run(path string, stage string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	reader, _ := utfbom.Skip(bytes.NewReader(data))
	buf, err := io.ReadAll(reader)
	if err != nil {
		return err
	}

	// This code is utter crap...
	content := strings.ReplaceAll(string(buf), "\r\n", "\n")
	content = strings.ReplaceAll(content, "\n", ";")
	content = strings.ReplaceAll(content, `"`, `\"`)

	log.Printf("Starting stage on %s %s (%s)", e.driver.GetMachineName(), stage, path)
	e.runCommand(content, true)
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

	output, err := client.Output(command)
	if printOutput {
		fmt.Printf("%s", output)
	}

	if err != nil {
		return fmt.Errorf(`ssh command error:
	command : %s
	err     : %v
	output  : %s`, command, err, output)
	}

	return nil
}
