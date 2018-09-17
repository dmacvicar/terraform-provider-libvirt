package libvirt

import (
	"os"
	"path/filepath"
	"testing"
)

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
	for _, file := range []string{"user-data", "meta-data", "network-config"} {
		check, err := exists(filepath.Join(dir, file))
		if !check {
			t.Errorf("%s not found: %v", file, err)
		}
	}
}

func TestCloudInitCreateISONoExternalTool(t *testing.T) {
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
