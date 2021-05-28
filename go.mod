module github.com/dmacvicar/terraform-provider-libvirt

require (
	github.com/c4milo/gotoolkit v0.0.0-20170704181456-e37eeabad07e // indirect
	github.com/coreos/go-semver v0.3.0 // indirect
	github.com/davecgh/go-spew v1.1.1
	github.com/digitalocean/go-libvirt v0.0.0-20210524223541-696696fc24e0
	github.com/google/go-cmp v0.5.1 // indirect
	github.com/google/uuid v1.1.2
	github.com/hashicorp/terraform-plugin-sdk v1.4.0
	github.com/hooklift/assert v0.0.0-20170704181755-9d1defd6d214 // indirect
	github.com/hooklift/iso9660 v1.0.0
	github.com/libvirt/libvirt-go v5.10.0+incompatible
	github.com/libvirt/libvirt-go-xml v5.10.0+incompatible
	github.com/mattn/go-colorable v0.1.8 // indirect
	github.com/mattn/goveralls v0.0.2
	github.com/pborman/uuid v1.2.0 // indirect
	github.com/stretchr/testify v1.6.1
	github.com/terraform-providers/terraform-provider-ignition v1.2.1
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9 // indirect
	golang.org/x/lint v0.0.0-20191125180803-fdd1cda4f05f
	golang.org/x/sys v0.0.0-20210514084401-e8d321eab015 // indirect
	google.golang.org/protobuf v1.25.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

replace git.apache.org/thrift.git => github.com/apache/thrift v0.0.0-20180902110319-2566ecd5d999

go 1.13
