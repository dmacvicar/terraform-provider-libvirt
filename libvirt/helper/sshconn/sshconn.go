package sshconn

import (
	"fmt"
	"io/ioutil"

	"github.com/mitchellh/go-homedir"
	"golang.org/x/crypto/ssh"
)

type SSHConn struct {
	Host    string
	Port    int
	User    string
	Keypath string
	Client  *ssh.Client
	Session *ssh.Session
}

func (s *SSHConn) Connect() error {
	var keypath string
	if s.Keypath == "" {
		homepath, err := homedir.Dir()
		if err != nil {
			return err
		}
		keypath = homepath + "/.ssh/id_rsa"
	} else {
		keypath = s.Keypath
	}

	key, err := ioutil.ReadFile(keypath)
	if err != nil {
		return err
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return err
	}

	config := &ssh.ClientConfig{}
	config.SetDefaults()
	config.User = s.User
	config.Auth = []ssh.AuthMethod{ssh.PublicKeys(signer)}
	config.HostKeyCallback = ssh.InsecureIgnoreHostKey()

	addr := fmt.Sprintf("%s:%d", s.Host, s.Port)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return err
	} else {
		s.Client = client
		return nil
	}
}

func (s *SSHConn) ExecCommand(cmd string) (string, error) {
	session, err := s.Client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	bs, err := session.CombinedOutput(cmd)
	if err != nil {
		return "", err
	}
	return string(bs), nil
}
