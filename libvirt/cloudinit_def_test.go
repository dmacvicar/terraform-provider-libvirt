package libvirt

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"gopkg.in/yaml.v2"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewCloudInitDef(t *testing.T) {
	ci := newCloudInitDef()

	if ci.Metadata.InstanceID == "" {
		t.Error("Expected metadata InstanceID not to be empty")
	}
}

func TestTerraformKeyOps(t *testing.T) {
	ci := newCloudInitDef()

	volKey := "volume-key"

	terraformId := ci.buildTerraformKey(volKey)
	if terraformId == "" {
		t.Error("key should not be empty")
	}

	actualKey, _ := getCloudInitVolumeKeyFromTerraformID(terraformId)
	if actualKey != volKey {
		t.Error("wrong key returned")
	}
}

func TestCreateFiles(t *testing.T) {
	ci := newCloudInitDef()

	dir, err := ci.createFiles()
	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}
	defer os.RemoveAll(dir)

	for _, file := range []string{USERDATA, METADATA} {
		check, err := exists(filepath.Join(dir, file))
		if !check {
			t.Errorf("%s not found: %v", file, err)
		}
	}
}

func TestCreateISONoExteralTool(t *testing.T) {
	path := os.Getenv("PATH")
	defer os.Setenv("PATH", path)

	os.Setenv("PATH", "/")

	ci := newCloudInitDef()

	iso, err := ci.createISO()
	if err == nil {
		t.Errorf("Expected error")
	}

	if iso != "" {
		t.Errorf("Expected iso to be empty")
	}
}

func TestConvertUserDataToMapPreservesCloudInitNames(t *testing.T) {
	ud := CloudInitUserData{
		SSHAuthorizedKeys: []string{"key1"},
	}

	actual, err := convertUserDataToMap(ud)
	if err != nil {
		t.Errorf("Unexpectd error %v", err)
	}

	_, ok := actual["ssh_authorized_keys"]
	if !ok {
		t.Error("Could not found ssh_authorized_keys key")
	}
}

func TestMergeEmptyUserDataIntoUserDataRaw(t *testing.T) {
	ud := CloudInitUserData{}

	var userDataRaw = `
new-key: new-value-set-by-extra
ssh_authorized_keys:
  - key2-from-extra-data
`

	res, err := mergeUserDataIntoUserDataRaw(ud, userDataRaw)
	if err != nil {
		t.Errorf("Unexpectd error %v", err)
	}

	actual := make(map[string]interface{})
	err = yaml.Unmarshal([]byte(res), &actual)
	if err != nil {
		t.Errorf("Unexpectd error %v", err)
	}

	if _, ok := actual["ssh_authorized_keys"]; !ok {
		t.Error("ssh_authorized_keys missing")
	}

	if _, ok := actual["new-key"]; !ok {
		t.Error("new-key missing")
	}
}

func TestMergeUserDataIntoEmptyUserDataRaw(t *testing.T) {
	ud := CloudInitUserData{
		SSHAuthorizedKeys: []string{"key1"},
	}
	var userDataRaw string

	res, err := mergeUserDataIntoUserDataRaw(ud, userDataRaw)
	if err != nil {
		t.Errorf("Unexpectd error %v", err)
	}

	actual := make(map[string]interface{})
	err = yaml.Unmarshal([]byte(res), &actual)
	if err != nil {
		t.Errorf("Unexpectd error %v", err)
	}

	if _, ok := actual["ssh_authorized_keys"]; !ok {
		t.Error("ssh_authorized_keys missing")
	}
}

func TestMergeUserDataIntoUserDataRawGivesPrecedenceToRawData(t *testing.T) {
	ud_key := "user-data-key"
	ud := CloudInitUserData{
		SSHAuthorizedKeys: []string{ud_key},
	}

	var userDataRaw = `
new-key: new-value-set-by-extra
ssh_authorized_keys:
  - key2-from-extra-data
`

	res, err := mergeUserDataIntoUserDataRaw(ud, userDataRaw)
	if err != nil {
		t.Errorf("Unexpectd error %v", err)
	}

	if strings.Contains(res, ud_key) {
		t.Error("Should not have found string defined by user data")
	}

	if !strings.Contains(res, "key2-from-extra-data") {
		t.Error("Should have found string defined by raw data")
	}
}

func TestCreateCloudIsoViaPlugin(t *testing.T) {
	var config = fmt.Sprintf(`

   resource "libvirt_cloudinit" "test" {
   name = "commoninit.iso"
   local_hostname = "tango"
   pool =           "default"
   user_data =      "#cloud-config\nssh_authorized_keys: []\n" 
    }
        `)
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(s *terraform.State) error {
			return nil
		},
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"libvirt_cloudinit.test", "name", "commoninit.iso"),
					resource.TestCheckResourceAttr(
						"libvirt_cloudinit.test", "local_hostname", "tango"),
				),
			},
		},
	})
}

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}
