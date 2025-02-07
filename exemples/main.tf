terraform {
  required_providers {
    windows-dsc = {
      source = "local/FranckSallet/windows-dsc"
    }
  }
}

provider "windows-dsc" {
  server_address       = "172.18.190.4"
  ssh_username         = "adminlocalecritel"
  ssh_private_key_path = "~/.ssh/id_rsa"
}

resource "windows-dsc_windowsfeature" "iis" {
  name                     = "Web-Server"
  ensure                   = "Present"
  include_all_sub_features = false                                  # Désactiver l'installation de toutes les sous-fonctionnalités
  sub_features             = ["Web-Common-Http", "Web-Default-Doc"] # Sous-fonctionnalités spécifiques
}
