module github.com/dmacvicar/terraform-provider-libvirt

require (
	github.com/c4milo/gotoolkit v0.0.0-20170704181456-e37eeabad07e // indirect
	github.com/community-terraform-providers/terraform-provider-ignition/v2 v2.1.2
	github.com/davecgh/go-spew v1.1.1
	github.com/digitalocean/go-libvirt v0.0.0-20220616141158-7ed4ed4decd9
	github.com/google/uuid v1.1.2
	github.com/hashicorp/terraform-plugin-sdk/v2 v2.21.0
	github.com/hooklift/assert v0.0.0-20170704181755-9d1defd6d214 // indirect
	github.com/hooklift/iso9660 v1.0.0
	github.com/mattn/goveralls v0.0.2
	github.com/pborman/uuid v1.2.0 // indirect
	github.com/stretchr/testify v1.7.2
	golang.org/x/crypto v0.0.0-20220517005047-85d78b3ac167
	golang.org/x/lint v0.0.0-20200302205851-738671d3881b
	golang.org/x/sys v0.0.0-20220517195934-5e4e11fc645e // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	libvirt.org/go/libvirtxml v1.8003.0
)

replace git.apache.org/thrift.git => github.com/apache/thrift v0.0.0-20180902110319-2566ecd5d999

replace golang.org/x/crypto v0.0.0-20220112180741-5e0467b6c7ce => github.com/dmacvicar/golang-x-crypto v0.0.0-20220126233154-a96af8f07497

go 1.18

replace github.com/community-terraform-providers/terraform-provider-ignition/v2 => github.com/dmacvicar/terraform-provider-ignition/v2 v2.1.3-0.20210701165004-13acf61ca184
