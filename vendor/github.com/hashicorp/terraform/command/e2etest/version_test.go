package e2etest

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hashicorp/terraform/e2e"
	tfcore "github.com/hashicorp/terraform/terraform"
)

func TestVersion(t *testing.T) {
	// Along with testing the "version" command in particular, this serves
	// as a good smoke test for whether the Terraform binary can even be
	// compiled and run, since it doesn't require any external network access
	// to do its job.

	t.Parallel()

	fixturePath := filepath.Join("test-fixtures", "empty")
	tf := e2e.NewBinary(terraformBin, fixturePath)
	defer tf.Close()

	stdout, stderr, err := tf.Run("version")
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	if stderr != "" {
		t.Errorf("unexpected stderr output:\n%s", stderr)
	}

	wantVersion := fmt.Sprintf("Terraform %s", tfcore.VersionString())
	if strings.Contains(stdout, wantVersion) {
		t.Errorf("output does not contain our current version %q:\n%s", wantVersion, stdout)
	}
}
