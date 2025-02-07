terraform {
  required_providers {
    windows-dsc = {
      source = "local/FranckSallet/windows-dsc"
    }
  }
}

provider "windows-dsc" {}

resource "windows-dsc_windowsfeature" "iis" {
  server_address       = "172.18.190.4"
  ssh_username         = "adminlocalecritel"
  ssh_private_key_path = "~/.ssh/id_rsa"

  name                     = "Telnet-Client"
  ensure                   = "Present"
  include_all_sub_features = false
  sub_features             = []
}
