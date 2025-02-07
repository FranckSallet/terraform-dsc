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
		Schema: map[string]*schema.Schema{}, // Aucun schéma nécessaire pour le provider
		ResourcesMap: map[string]*schema.Resource{
			"windows-dsc_windowsfeature": resources.WindowsFeature(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	// Le provider n'a plus besoin de configurer les paramètres SSH
	return nil, nil
}

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: Provider,
	})
}
