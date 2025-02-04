package resources

import (
	"fmt"

	"golang.org/x/crypto/ssh"
)

// SSHClient représente un client SSH connecté à un serveur distant
type SSHClient struct {
	Client *ssh.Client
}

// NewSSHClient crée une nouvelle connexion SSH
func NewSSHClient(address, username, password string) (*SSHClient, error) {
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	client, err := ssh.Dial("tcp", address+":22", config)
	if err != nil {
		return nil, fmt.Errorf("échec de la connexion SSH : %v", err)
	}

	return &SSHClient{Client: client}, nil
}

// RunCommand exécute une commande sur le serveur distant
func (c *SSHClient) RunCommand(command string) (string, error) {
	session, err := c.Client.NewSession()
	if err != nil {
		return "", fmt.Errorf("échec de la création de la session SSH : %v", err)
	}
	defer session.Close()

	output, err := session.Output(command)
	if err != nil {
		return "", fmt.Errorf("échec de l'exécution de la commande : %v", err)
	}

	return string(output), nil
}

// Close ferme la connexion SSH
func (c *SSHClient) Close() error {
	return c.Client.Close()
}
