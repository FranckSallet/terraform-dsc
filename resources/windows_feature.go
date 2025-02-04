package resources

import (
	"context"
	"fmt"
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
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true, // Le nom ne peut pas être modifié après la création
			},
			"ensure": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "Present",
				ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
					v := val.(string)
					if v != "Present" && v != "Absent" {
						errs = append(errs, fmt.Errorf("%s doit être 'Present' ou 'Absent'", key))
					}
					return
				},
			},
			"include_all_sub_features": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"server_address": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true, // L'adresse du serveur ne peut pas être modifiée après la création
			},
			"ssh_username": {
				Type:     schema.TypeString,
				Required: true,
			},
			"ssh_password": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},
		},
	}
}

// Fonction pour créer une ressource
func WindowsFeatureCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Connexion SSH
	sshClient, err := NewSSHClient(
		d.Get("server_address").(string),
		d.Get("ssh_username").(string),
		d.Get("ssh_password").(string),
	)
	if err != nil {
		return diag.FromErr(fmt.Errorf("échec de la connexion SSH : %v", err))
	}
	defer sshClient.Close()

	// Vérifier si la fonctionnalité existe
	output, err := sshClient.RunCommand("powershell -Command \"Get-WindowsFeature -Name " + d.Get("name").(string) + "\"")
	if err != nil {
		return diag.FromErr(fmt.Errorf("échec de la vérification de la fonctionnalité : %v", err))
	}
	if strings.Contains(output, "FeatureNotFound") {
		return diag.FromErr(fmt.Errorf("la fonctionnalité '%s' n'existe pas", d.Get("name").(string)))
	}

	// Script PowerShell pour appliquer DSC
	script := fmt.Sprintf(`
        Configuration ConfigureFeature {
            Import-DscResource -ModuleName PSDesiredStateConfiguration
            Node "localhost" {
                WindowsFeature %s {
                    Name                 = "%s"
                    Ensure               = "%s"
                    IncludeAllSubFeature = %t
                }
            }
        }
        ConfigureFeature
        Start-DscConfiguration -Path .\ConfigureFeature -Wait -Verbose -Force
    `, d.Get("name").(string), d.Get("name").(string), d.Get("ensure").(string), d.Get("include_all_sub_features").(bool))

	// Exécution du script
	_, err = sshClient.RunCommand("powershell -Command \"" + script + "\"")
	if err != nil {
		return diag.FromErr(fmt.Errorf("échec de l'exécution du script PowerShell : %v", err))
	}

	// Définition de l'ID de la ressource
	d.SetId(d.Get("name").(string))

	return WindowsFeatureRead(ctx, d, meta)
}

// Fonction pour lire l'état d'une ressource
func WindowsFeatureRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Connexion SSH
	sshClient, err := NewSSHClient(
		d.Get("server_address").(string),
		d.Get("ssh_username").(string),
		d.Get("ssh_password").(string),
	)
	if err != nil {
		return diag.FromErr(fmt.Errorf("échec de la connexion SSH : %v", err))
	}
	defer sshClient.Close()

	// Vérification de l'état de la fonctionnalité
	output, err := sshClient.RunCommand("powershell -Command \"Get-WindowsFeature -Name " + d.Get("name").(string) + " | Select-Object -Property Installed\"")
	if err != nil {
		return diag.FromErr(fmt.Errorf("échec de la vérification de l'état de la fonctionnalité : %v", err))
	}

	// Analyse de la sortie
	if strings.Contains(output, "True") {
		d.Set("ensure", "Present")
	} else {
		d.Set("ensure", "Absent")
	}

	return nil
}

// Fonction pour mettre à jour une ressource
func WindowsFeatureUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Si le nom ou l'adresse du serveur change, recréer la ressource
	if d.HasChange("name") || d.HasChange("server_address") {
		return WindowsFeatureCreate(ctx, d, meta)
	}

	// Sinon, appliquer la mise à jour via DSC
	return WindowsFeatureCreate(ctx, d, meta)
}

// Fonction pour supprimer une ressource
func WindowsFeatureDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Connexion SSH
	sshClient, err := NewSSHClient(
		d.Get("server_address").(string),
		d.Get("ssh_username").(string),
		d.Get("ssh_password").(string),
	)
	if err != nil {
		return diag.FromErr(fmt.Errorf("échec de la connexion SSH : %v", err))
	}
	defer sshClient.Close()

	// Script PowerShell pour supprimer la fonctionnalité
	script := fmt.Sprintf(`
        Configuration RemoveFeature {
            Import-DscResource -ModuleName PSDesiredStateConfiguration
            Node "localhost" {
                WindowsFeature %s {
                    Name   = "%s"
                    Ensure = "Absent"
                }
            }
        }
        RemoveFeature
        Start-DscConfiguration -Path .\RemoveFeature -Wait -Verbose -Force
    `, d.Get("name").(string), d.Get("name").(string))

	// Exécution du script
	_, err = sshClient.RunCommand("powershell -Command \"" + script + "\"")
	if err != nil {
		return diag.FromErr(fmt.Errorf("échec de la suppression de la fonctionnalité : %v", err))
	}

	// Suppression de l'ID de la ressource
	d.SetId("")

	return nil
}
