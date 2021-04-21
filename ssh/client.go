package ssh

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/docker/machine/libmachine/mcnutils"
	"github.com/moby/term"
	"github.com/prometheus/common/log"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
)

type Auth struct {
	Passwords []string
	Keys      []string
}

type Client interface {
	Output(command string) (string, error)
	OutputWithPty(command string) (string, error)
	Shell(args ...string) error

	// Start starts the specified command without waiting for it to finish. You
	// have to call the Wait function for that.
	//
	// The first two io.ReadCloser are the standard output and the standard
	// error of the executing command respectively. The returned error follows
	// the same logic as in the exec.Cmd.Start function.
	Start(command string) (io.ReadCloser, io.ReadCloser, error)

	// Wait waits for the command started by the Start function to exit. The
	// returned error follows the same logic as in the exec.Cmd.Wait function.
	Wait() error
}

type NativeClient struct {
	Config      ssh.ClientConfig
	Hostname    string
	Port        int
	openSession *ssh.Session
	openClient  *ssh.Client
}

func NewClient(user string, host string, port int, auth *Auth) (Client, error) {
	log.Debug("Using SSH client type: native")
	client, err := NewNativeClient(user, host, port, auth)
	log.Debug(client)
	return client, err
}

func NewNativeClient(user, host string, port int, auth *Auth) (Client, error) {
	config, err := NewNativeConfig(user, auth)
	if err != nil {
		return nil, fmt.Errorf("Error getting config for native Go SSH: %s", err)
	}

	return &NativeClient{
		Config:   config,
		Hostname: host,
		Port:     port,
	}, nil
}

func NewNativeConfig(user string, auth *Auth) (ssh.ClientConfig, error) {
	var (
		authMethods []ssh.AuthMethod
	)

	for _, k := range auth.Keys {
		key, err := ioutil.ReadFile(k)
		if err != nil {
			return ssh.ClientConfig{}, err
		}

		privateKey, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return ssh.ClientConfig{}, err
		}

		authMethods = append(authMethods, ssh.PublicKeys(privateKey))
	}

	for _, p := range auth.Passwords {
		authMethods = append(authMethods, ssh.Password(p))
	}

	return ssh.ClientConfig{
		User:            user,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}, nil
}

func (client *NativeClient) Wait() error {
	err := client.openSession.Wait()
	if err != nil {
		return err
	}

	_ = client.openSession.Close()

	err = client.openClient.Close()
	if err != nil {
		return err
	}

	client.openSession = nil
	client.openClient = nil
	return nil
}

func (client *NativeClient) Shell(args ...string) error {
	var (
		termWidth, termHeight int
	)
	conn, err := ssh.Dial("tcp", net.JoinHostPort(client.Hostname, strconv.Itoa(client.Port)), &client.Config)
	if err != nil {
		return err
	}
	defer closeConn(conn)

	session, err := conn.NewSession()
	if err != nil {
		return err
	}

	defer session.Close()

	session.Stdout = os.Stdout
	session.Stderr = os.Stderr
	session.Stdin = os.Stdin

	modes := ssh.TerminalModes{
		ssh.ECHO: 1,
	}

	fd := os.Stdin.Fd()

	if term.IsTerminal(fd) {
		oldState, err := term.MakeRaw(fd)
		if err != nil {
			return err
		}

		defer term.RestoreTerminal(fd, oldState)

		winsize, err := term.GetWinsize(fd)
		if err != nil {
			termWidth = 80
			termHeight = 24
		} else {
			termWidth = int(winsize.Width)
			termHeight = int(winsize.Height)
		}
	}

	if err := session.RequestPty("xterm", termHeight, termWidth, modes); err != nil {
		return err
	}

	if len(args) == 0 {
		if err := session.Shell(); err != nil {
			return err
		}
		if err := session.Wait(); err != nil {
			return err
		}
	} else {
		if err := session.Run(strings.Join(args, " ")); err != nil {
			return err
		}
	}
	return nil
}

func (client *NativeClient) Start(command string) (io.ReadCloser, io.ReadCloser, error) {
	conn, session, err := client.session(command)
	if err != nil {
		return nil, nil, err
	}

	stdout, err := session.StdoutPipe()
	if err != nil {
		return nil, nil, err
	}
	stderr, err := session.StderrPipe()
	if err != nil {
		return nil, nil, err
	}
	if err := session.Start(command); err != nil {
		return nil, nil, err
	}

	client.openClient = conn
	client.openSession = session
	return ioutil.NopCloser(stdout), ioutil.NopCloser(stderr), nil
}

func (client *NativeClient) Output(command string) (string, error) {
	conn, session, err := client.session(command)
	if err != nil {
		return "", nil
	}
	defer closeConn(conn)
	defer session.Close()

	output, err := session.CombinedOutput(command)

	return string(output), err
}

func (client *NativeClient) OutputWithPty(command string) (string, error) {
	conn, session, err := client.session(command)
	if err != nil {
		return "", nil
	}
	defer closeConn(conn)
	defer session.Close()

	fd := int(os.Stdout.Fd())

	termWidth, termHeight, err := terminal.GetSize(fd)
	if err != nil {
		return "", err
	}

	modes := ssh.TerminalModes{
		ssh.ECHO:          0,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}

	// request tty -- fixes error with hosts that use
	// "Defaults requiretty" in /etc/sudoers - I'm looking at you RedHat
	if err := session.RequestPty("xterm", termHeight, termWidth, modes); err != nil {
		return "", err
	}

	output, err := session.CombinedOutput(command)

	return string(output), err
}

func (client *NativeClient) dialSuccess() bool {
	conn, err := ssh.Dial("tcp", net.JoinHostPort(client.Hostname, strconv.Itoa(client.Port)), &client.Config)
	if err != nil {
		log.Debugf("Error dialing TCP: %s", err)
		return false
	}
	closeConn(conn)
	return true
}

func (client *NativeClient) session(command string) (*ssh.Client, *ssh.Session, error) {
	if err := mcnutils.WaitFor(client.dialSuccess); err != nil {
		return nil, nil, fmt.Errorf("Error attempting SSH client dial: %s", err)
	}

	conn, err := ssh.Dial("tcp", net.JoinHostPort(client.Hostname, strconv.Itoa(client.Port)), &client.Config)
	if err != nil {
		return nil, nil, fmt.Errorf("Mysterious error dialing TCP for SSH (we already succeeded at least once) : %s", err)
	}
	session, err := conn.NewSession()

	return conn, session, err
}

func closeConn(c io.Closer) {
	err := c.Close()
	if err != nil {
		log.Debugf("Error closing SSH Client: %s", err)
	}
}
