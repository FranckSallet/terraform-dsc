package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// WindowsFeature définit la ressource Terraform
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

// Fonction pour créer une ressource
func WindowsFeatureCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Récupère les paramètres du provider
	providerConfig := meta.(map[string]string)
	serverAddress := providerConfig["server_address"]
	sshUsername := providerConfig["ssh_username"]
	sshPassword := providerConfig["ssh_password"]

	// Connexion SSH
	sshClient, err := NewSSHClient(serverAddress, sshUsername, sshPassword)
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

	// Récupérer les sous-fonctionnalités spécifiées
	subFeatures := d.Get("sub_features").([]interface{})
	subFeaturesList := make([]string, len(subFeatures))
	for i, v := range subFeatures {
		subFeaturesList[i] = v.(string)
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
    `, d.Get("name").(string), d.Get("name").(string), d.Get("ensure").(string), d.Get("include_all_sub_features").(bool))

	// Ajouter les sous-fonctionnalités spécifiées si 'include_all_sub_features' est false
	if len(subFeaturesList) > 0 && !d.Get("include_all_sub_features").(bool) {
		script += "SubFeatures = @("
		for _, subFeature := range subFeaturesList {
			script += fmt.Sprintf("\"%s\", ", subFeature)
		}
		script = script[:len(script)-2] // Supprimer la dernière virgule et l'espace
		script += ")\n"
	}

	script += `
                }
            }
        }
        ConfigureFeature
        Start-DscConfiguration -Path .\ConfigureFeature -Wait -Verbose -Force
    `

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
	// Récupère les paramètres du provider
	providerConfig := meta.(map[string]string)
	serverAddress := providerConfig["server_address"]
	sshUsername := providerConfig["ssh_username"]
	sshPassword := providerConfig["ssh_password"]

	// Connexion SSH
	sshClient, err := NewSSHClient(serverAddress, sshUsername, sshPassword)
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
	// Si le nom ou les sous-fonctionnalités changent, recréer la ressource
	if d.HasChange("name") || d.HasChange("sub_features") || d.HasChange("include_all_sub_features") {
		return WindowsFeatureCreate(ctx, d, meta)
	}

	// Sinon, appliquer la mise à jour via DSC
	return WindowsFeatureCreate(ctx, d, meta)
}

// Fonction pour supprimer une ressource
func WindowsFeatureDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Récupère les paramètres du provider
	providerConfig := meta.(map[string]string)
	serverAddress := providerConfig["server_address"]
	sshUsername := providerConfig["ssh_username"]
	sshPassword := providerConfig["ssh_password"]

	// Connexion SSH
	sshClient, err := NewSSHClient(serverAddress, sshUsername, sshPassword)
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
