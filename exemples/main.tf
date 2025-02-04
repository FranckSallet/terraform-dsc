terraform {
  required_providers {
    windows-dsc = {
      source  = "local/fsallet/windows-dsc"
      version = "1.0.0"
    }
  }
}

provider "windows-dsc" {
  server_address = "192.168.1.100" # Adresse IP ou nom d'hôte du serveur Windows
  ssh_username   = "admin"         # Nom d'utilisateur SSH
  ssh_password   = "password"      # Mot de passe SSH
}

resource "windows-dsc_windowsfeature" "iis" {
  name                    = "Web-Server"
  ensure                  = "Present"
  include_all_sub_features = true
}