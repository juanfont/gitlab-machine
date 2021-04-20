package executor

import (
	"fmt"
	"log"

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

	return nil
}

// Run executes the required script
func (e *Executor) Run() {

}

// Cleanup releases the resources once the job has finished
func (e *Executor) CleanUp() {

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
