# Build
go mod init github.com/FranckSallet/windows-dsc
go get github.com/hashicorp/terraform-plugin-sdk/v2@v2.26.1
go get golang.org/x/crypto@v0.14.0
go mod tidy
go build -o ~/.terraform.d/plugins/local/FranckSallet/windows-dsc/1.0.0/linux_amd64/terraform-provider-windows-dsc