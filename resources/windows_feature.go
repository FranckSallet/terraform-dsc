package resources

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func WindowsFeature() *schema.Resource {
	return &schema.Resource{
		CreateContext: WindowsFeatureCreate,
		ReadContext:   WindowsFeatureRead,
		UpdateContext: WindowsFeatureUpdate,
		DeleteContext: WindowsFeatureDelete,
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
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true, // Le nom ne peut pas être modifié après la création
			},
			"include_all_sub_features": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Si true, toutes les sous-fonctionnalités seront installées. Si false, seules les sous-fonctionnalités spécifiées dans 'sub_features' seront installées.",
			},
			"sub_features": {
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Liste des sous-fonctionnalités à installer. Ignoré si 'include_all_sub_features' est true.",
			},
		},
	}
}

func WindowsFeatureCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	serverAddress := d.Get("server_address").(string)
	sshUsername := d.Get("ssh_username").(string)
	sshPassword := d.Get("ssh_password").(string)
	sshPrivateKeyPath := d.Get("ssh_private_key_path").(string)

	sshClient, err := NewSSHClient(serverAddress, sshUsername, sshPassword, sshPrivateKeyPath)
	if err != nil {
		return diag.FromErr(fmt.Errorf("échec de la connexion SSH : %v", err))
	}
	defer sshClient.Close()

	output, err := sshClient.RunCommand("powershell -Command \"Get-WindowsFeature -Name " + d.Get("name").(string) + "\"")
	if err != nil {
		return diag.FromErr(fmt.Errorf("échec de la vérification de la fonctionnalité : %v", err))
	}
	if strings.Contains(output, "FeatureNotFound") {
		return diag.FromErr(fmt.Errorf("la fonctionnalité '%s' n'existe pas", d.Get("name").(string)))
	}

	subFeatures := d.Get("sub_features").([]interface{})
	subFeaturesList := make([]string, len(subFeatures))
	for i, v := range subFeatures {
		subFeaturesList[i] = v.(string)
	}

	// Le paramètre "ensure" est supprimé, car il est implicite dans Terraform
	script := fmt.Sprintf(`
        Configuration ConfigureFeature {
            Import-DscResource -ModuleName PSDesiredStateConfiguration
            Node "localhost" {
                WindowsFeature %s {
                    Name                 = "%s"
                    Ensure               = "Present" // Toujours "Present" car la ressource est déclarée
                    IncludeAllSubFeature = %t
    `, d.Get("name").(string), d.Get("name").(string), d.Get("include_all_sub_features").(bool))

	if len(subFeaturesList) > 0 && !d.Get("include_all_sub_features").(bool) {
		script += "SubFeatures = @("
		for _, subFeature := range subFeaturesList {
			script += fmt.Sprintf("\"%s\", ", subFeature)
		}
		script = script[:len(script)-2]
		script += ")\n"
	}

	script += `
                }
            }
        }
        ConfigureFeature
        Start-DscConfiguration -Path .\ConfigureFeature -Wait -Verbose -Force
    `

	_, err = sshClient.RunCommand("powershell -Command \"" + script + "\"")
	if err != nil {
		return diag.FromErr(fmt.Errorf("échec de l'exécution du script PowerShell : %v", err))
	}

	d.SetId(d.Get("name").(string) + "@" + serverAddress)

	return WindowsFeatureRead(ctx, d, meta)
}

func WindowsFeatureRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	serverAddress := d.Get("server_address").(string)
	sshUsername := d.Get("ssh_username").(string)
	sshPassword := d.Get("ssh_password").(string)
	sshPrivateKeyPath := d.Get("ssh_private_key_path").(string)

	sshClient, err := NewSSHClient(serverAddress, sshUsername, sshPassword, sshPrivateKeyPath)
	if err != nil {
		return diag.FromErr(fmt.Errorf("échec de la connexion SSH : %v", err))
	}
	defer sshClient.Close()

	// Exécuter la commande pour vérifier l'état de la fonctionnalité
	output, err := sshClient.RunCommand("powershell -Command \"Get-WindowsFeature -Name " + d.Get("name").(string) + " | Select-Object -Property Installed\"")
	if err != nil {
		return diag.FromErr(fmt.Errorf("échec de la vérification de l'état de la fonctionnalité : %v", err))
	}

	// Vérifier si la fonctionnalité est installée et logger le résultat
	if strings.Contains(output, "True") {
		log.Printf("La fonctionnalité '%s' est installée sur le serveur %s\n", d.Get("name").(string), serverAddress)
	} else {
		log.Printf("La fonctionnalité '%s' n'est pas installée sur le serveur %s\n", d.Get("name").(string), serverAddress)
	}

	return nil
}

func WindowsFeatureUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if d.HasChange("name") || d.HasChange("sub_features") || d.HasChange("include_all_sub_features") {
		return WindowsFeatureCreate(ctx, d, meta)
	}

	return WindowsFeatureCreate(ctx, d, meta)
}

func WindowsFeatureDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	serverAddress := d.Get("server_address").(string)
	sshUsername := d.Get("ssh_username").(string)
	sshPassword := d.Get("ssh_password").(string)
	sshPrivateKeyPath := d.Get("ssh_private_key_path").(string)

	sshClient, err := NewSSHClient(serverAddress, sshUsername, sshPassword, sshPrivateKeyPath)
	if err != nil {
		return diag.FromErr(fmt.Errorf("échec de la connexion SSH : %v", err))
	}
	defer sshClient.Close()

	script := fmt.Sprintf(`
        Configuration RemoveFeature {
            Import-DscResource -ModuleName PSDesiredStateConfiguration
            Node "localhost" {
                WindowsFeature %s {
                    Name   = "%s"
                    Ensure = "Absent" // Toujours "Absent" lors de la suppression
                }
            }
        }
        RemoveFeature
        Start-DscConfiguration -Path .\RemoveFeature -Wait -Verbose -Force
    `, d.Get("name").(string), d.Get("name").(string))

	_, err = sshClient.RunCommand("powershell -Command \"" + script + "\"")
	if err != nil {
		return diag.FromErr(fmt.Errorf("échec de la suppression de la fonctionnalité : %v", err))
	}

	d.SetId("")

	return nil
}
