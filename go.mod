module github.com/dmacvicar/terraform-provider-libvirt

require (
	github.com/c4milo/gotoolkit v0.0.0-20170704181456-e37eeabad07e // indirect
	github.com/coreos/go-semver v0.3.0 // indirect
	github.com/davecgh/go-spew v1.1.1
	github.com/digitalocean/go-libvirt v0.0.0-20210723161134-761cfeeb5968
	github.com/fatih/color v1.10.0 // indirect
	github.com/google/uuid v1.1.2
	github.com/hashicorp/terraform-plugin-sdk v1.9.0
	github.com/hooklift/assert v0.0.0-20170704181755-9d1defd6d214 // indirect
	github.com/hooklift/iso9660 v1.0.0
	github.com/libvirt/libvirt-go-xml v7.4.0+incompatible
	github.com/mattn/goveralls v0.0.2
	github.com/pborman/uuid v1.2.0 // indirect
	github.com/stretchr/testify v1.7.0
	github.com/terraform-providers/terraform-provider-ignition v1.2.1
	github.com/ulikunitz/xz v0.5.8 // indirect
	golang.org/x/crypto v0.0.0-20220112180741-5e0467b6c7ce
	golang.org/x/lint v0.0.0-20200302205851-738671d3881b
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

replace git.apache.org/thrift.git => github.com/apache/thrift v0.0.0-20180902110319-2566ecd5d999

go 1.13
