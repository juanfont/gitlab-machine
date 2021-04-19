package drivers

type Driver interface {
	Create(instanceName string) error
	RunCommand(instanceName string) error
	Destroy(instanceName string) error
}
