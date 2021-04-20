package drivers

import "github.com/juanfont/gitlab-machine/ssh"

type Driver interface {
	Create() error
	Destroy() error
	GetMachineName() string
	GetSSHClientFromDriver() (ssh.Client, error)
}
