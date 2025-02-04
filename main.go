package main

import (
	"github.com/fsallet/terraform-dsc/resources"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		ResourcesMap: map[string]*schema.Resource{
			"windowsdsc_feature": resources.WindowsFeature(), // Utilise la fonction WindowsFeature du package resources
		},
	}
}

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: Provider,
	})
}
