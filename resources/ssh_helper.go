package resources

import (
	"fmt"
	"io/ioutil"
	"log"

	"golang.org/x/crypto/ssh"
)

// SSHClient représente un client SSH connecté à un serveur distant
type SSHClient struct {
	Client *ssh.Client
}

// NewSSHClient crée une nouvelle connexion SSH avec authentification par mot de passe ou clé SSH
func NewSSHClient(address, username, password, privateKeyPath string) (*SSHClient, error) {
	// Configuration SSH de base
	config := &ssh.ClientConfig{
		User:            username,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // Attention : désactive la vérification de la clé hôte (à utiliser avec précaution)
	}

	// Authentification par clé SSH
	if privateKeyPath != "" {
		key, err := ioutil.ReadFile(privateKeyPath)
		if err != nil {
			return nil, fmt.Errorf("échec de la lecture de la clé privée : %v", err)
		}

		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return nil, fmt.Errorf("échec du parsing de la clé privée : %v", err)
		}

		config.Auth = []ssh.AuthMethod{
			ssh.PublicKeys(signer), // Utilisation de la clé privée pour l'authentification
		}
	} else if password != "" {
		// Authentification par mot de passe
		config.Auth = []ssh.AuthMethod{
			ssh.Password(password),
		}
	} else {
		return nil, fmt.Errorf("aucune méthode d'authentification fournie (mot de passe ou clé SSH)")
	}

	// Connexion SSH
	client, err := ssh.Dial("tcp", address+":22", config)
	if err != nil {
		return nil, fmt.Errorf("échec de la connexion SSH : %v", err)
	}

	log.Printf("Connexion SSH établie avec succès à %s\n", address)
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

	log.Printf("Commande exécutée avec succès : %s\n", command)
	return string(output), nil
}

// Close ferme la connexion SSH
func (c *SSHClient) Close() error {
	if c.Client != nil {
		log.Println("Fermeture de la connexion SSH")
		return c.Client.Close()
	}
	return nil
}
