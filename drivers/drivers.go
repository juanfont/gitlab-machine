package drivers

type Driver interface {
	Create(instanceName string) error
	RunCommand(instanceName string) (string, string)
	Destroy(instanceName string) error
}

type Instance interface {
}
