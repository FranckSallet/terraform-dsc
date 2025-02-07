package main

import (
	"context"

	"github.com/FranckSallet/windows-dsc/resources"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"server_address": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Adresse IP ou nom d'hôte du serveur Windows",
			},
			"ssh_username": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Nom d'utilisateur pour la connexion SSH",
			},
			"ssh_password": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Mot de passe pour la connexion SSH (ignoré si ssh_private_key_path est fourni)",
			},
			"ssh_private_key_path": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Chemin vers la clé privée SSH (prioritaire sur ssh_password)",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"terraform-dsc_windowsfeature": resources.WindowsFeature(),
		},
		ConfigureContextFunc: providerConfigure, // Fonction pour configurer le provider
	}
}

// providerConfigure configure le provider avec les paramètres fournis
func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	serverAddress := d.Get("server_address").(string)
	sshUsername := d.Get("ssh_username").(string)
	sshPassword := d.Get("ssh_password").(string)
	sshPrivateKeyPath := d.Get("ssh_private_key_path").(string)

	// Retourne les paramètres du provider pour qu'ils soient utilisés dans les ressources
	return map[string]string{
		"server_address":       serverAddress,
		"ssh_username":         sshUsername,
		"ssh_password":         sshPassword,
		"ssh_private_key_path": sshPrivateKeyPath,
	}, nil
}

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: Provider,
	})
}
