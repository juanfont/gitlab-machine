package executor

import "github.com/juanfont/gitlab-windows-custom-executor/drivers"

type Executor struct {
	driver *drivers.Driver
}

runner-$CUSTOM_ENV_CI_RUNNER_ID-project-$CUSTOM_ENV_CI_PROJECT_ID-concurrent-$CUSTOM_ENV_CI_CONCURRENT_PROJECT_ID-job-$CUSTOM_ENV_CI_JOB_ID

func NewExecutor(d *drivers.Driver) (*Executor, error) {
	e := Executor{}
	e.driver = d
	return &e, nil
}

// Prepare calls the driver to ready up a new execution environment
func (e *Executor) Prepare() {

}

// Run executes the required script
func (e *Executor) Run() {

}

// Cleanup releases the resources once the job has finished
func (e *Executor) CleanUp() {

}
