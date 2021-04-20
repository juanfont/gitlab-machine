package executor

import (
	"fmt"
	"log"
	"os"

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

	e.runCommands([]string{"choco install -y git git-lfs gitlab-runner"})

	return nil
}

// Run executes the required script
func (e *Executor) Run(path string, stage string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	log.Printf("Starting stage on %s %s", e.driver.GetMachineName(), stage)
	e.runCommands([]string{string(content)})
	return nil
}

// Cleanup releases the resources once the job has finished
func (e *Executor) CleanUp() error {
	err := e.driver.Destroy()
	return err
}

func (e *Executor) runCommands(commands []string) error {
	client, err := e.driver.GetSSHClientFromDriver()
	if err != nil {
		return err
	}
	for _, command := range commands {
		log.Printf("About to run SSH command:\n%s", command)
		output, err := client.Output(command)
		log.Printf("SSH cmd err, output: %v: %s", err, output)
		if err != nil {
			return fmt.Errorf(`ssh command error:
	command : %s
	err     : %v
	output  : %s`, command, err, output)
		}

	}
	return nil
}
