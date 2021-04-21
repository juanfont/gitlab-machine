package drivers

import "github.com/juanfont/gitlab-machine/ssh"

type OStype string

const (
	Windows OStype = "windows"
	Linux   OStype = "linux"
)

type Driver interface {
	Create() error
	Destroy() error
	GetMachineName() string
	GetOS() (OStype, error)
	GetSSHClientFromDriver() (ssh.Client, error)
}
