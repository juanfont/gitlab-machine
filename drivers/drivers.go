package drivers

import "github.com/juanfont/gitlab-machine/ssh"

type Driver interface {
	Create() error
	Destroy() error
	GetSSHClientFromDriver() (ssh.Client, error)
}
