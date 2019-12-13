package ignition

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"

	"github.com/coreos/ignition/config/validate/report"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		DataSourcesMap: map[string]*schema.Resource{
			"ignition_config":        dataSourceConfig(),
			"ignition_disk":          dataSourceDisk(),
			"ignition_raid":          dataSourceRaid(),
			"ignition_filesystem":    dataSourceFilesystem(),
			"ignition_file":          dataSourceFile(),
			"ignition_directory":     dataSourceDirectory(),
			"ignition_link":          dataSourceLink(),
			"ignition_systemd_unit":  dataSourceSystemdUnit(),
			"ignition_networkd_unit": dataSourceNetworkdUnit(),
			"ignition_user":          dataSourceUser(),
			"ignition_group":         dataSourceGroup(),
		},
	}
}

func id(input interface{}) string {
	b, _ := json.Marshal(input)
	return hash(string(b))
}

func hash(s string) string {
	sha := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sha[:])
}

func getInt(d *schema.ResourceData, key string) *int {
	var i *int
	if value, ok := d.GetOk(key); ok {
		n := value.(int)
		i = &n
	}

	return i
}

func handleReport(r report.Report) error {
	for _, e := range r.Entries {
		debug(e.String())
	}

	if r.IsFatal() {
		return fmt.Errorf("invalid configuration:\n%s", r.String())
	}

	return nil
}

func debug(format string, a ...interface{}) {
	log.Printf("[DEBUG] %s", fmt.Sprintf(format, a...))
}
