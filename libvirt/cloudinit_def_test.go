package libvirt

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v2"
)

func TestNewCloudInitDef(t *testing.T) {
	ci := newCloudInitDef()

	if ci.MetaData.InstanceID == "" {
		t.Error("Expected metadata InstanceID not to be empty")
	}
}

func TestCloudInitTerraformKeyOps(t *testing.T) {
	ci := newCloudInitDef()

	volKey := "volume-key"

	terraformID := ci.buildTerraformKey(volKey)
	if terraformID == "" {
		t.Error("key should not be empty")
	}

	actualKey, _ := getCloudInitVolumeKeyFromTerraformID(terraformID)
	if actualKey != volKey {
		t.Error("wrong key returned")
	}
}

func TestCloudInitCreateFiles(t *testing.T) {
	ci := newCloudInitDef()

	dir, err := ci.createFiles()
	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}
	defer os.RemoveAll(dir)

	for _, file := range []string{userData, metaData} {
		check, err := exists(filepath.Join(dir, file))
		if !check {
			t.Errorf("%s not found: %v", file, err)
		}
	}
}

func TestCloudInitCreateISONoExteralTool(t *testing.T) {
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

func TestCloudInitConvertUserDataToMapPreservesCloudInitNames(t *testing.T) {
	ud := defCloudInitUserData{
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

func TestCloudInitMergeEmptyUserDataIntoUserDataRaw(t *testing.T) {
	ud := defCloudInitUserData{}

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

func TestCloudInitMergeUserDataIntoEmptyUserDataRaw(t *testing.T) {
	ud := defCloudInitUserData{
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

func TestCloudInitMergeUserDataIntoUserDataRawGivesPrecedenceToRawData(t *testing.T) {
	udKey := "user-data-key"
	ud := defCloudInitUserData{
		SSHAuthorizedKeys: []string{udKey},
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

	if strings.Contains(res, udKey) {
		t.Error("Should not have found string defined by user data")
	}

	if !strings.Contains(res, "key2-from-extra-data") {
		t.Error("Should have found string defined by raw data")
	}
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
