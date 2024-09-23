package libvirt

import (
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// mockCapabilitiesXML provides a sample XML for libvirt capabilities.
const mockCapabilitiesXML = `
<capabilities>
	<host>
		<uuid>test-uuid</uuid>
		<cpu>
			<arch>x86_64</arch>
			<model>model-name</model>
			<vendor>vendor-name</vendor>
			<topology sockets="2" dies="1" cores="4" threads="2"/>
			<feature name="feature1"/>
			<feature name="feature2"/>
			<feature name="feature3"/>
		</cpu>
		<migration_features>
			<live />
			<uri_transports>
				<uri_transport>transport#1</uri_transport>
				<uri_transport>transport#2</uri_transport>
				<uri_transport>transport#3</uri_transport>
			</uri_transports>
		</migration_features>
	</host>
</capabilities>
`
// Mock implementation of the libvirt client.
type MockLibvirtClient struct{}
// Implement the ConnectGetCapabilities method.
func (m *MockLibvirtClient) ConnectGetCapabilities() (string, error) {
	return mockCapabilitiesXML, nil
}

func Test_datasourceLibvirtCapabilitiesRead(t *testing.T) {
	// Create a mock libvirt client.
	mockLibvirtClient := &MockLibvirtClient{}

	// Initialize a Terraform resource data schema
	d := schema.TestResourceDataRaw(t, datasourceLibvirtCapabilities().Schema, map[string]interface{}{})

	// Call the function under test
	err := getCapabilities(d, mockLibvirtClient)

	if err != nil {
		t.Fatalf("resourceLibvirtHostCapabilitiesRead() error = %v", err)
	}

	// Define expected values based on mockCapabilitiesXML
	expectedUUID := "test-uuid"
	expectedArch := "x86_64"
	expectedModel := "model-name"
	expectedVendor := "vendor-name"
	expectedSockets := 2
	expectedDies := 1
	expectedCores := 4
	expectedThreads := 2
	expectedLiveMigrationSupport := true
	expectedLiveMigrationTransports := []interface{}{"transport#1", "transport#2", "transport#3"}
	expectedFeatures := []interface{}{"feature1", "feature2", "feature3"}

	// Assert that the resource data is set correctly
	if got := d.Get("arch"); got != expectedArch {
		t.Errorf("resourceLibvirtNodeInfoRead() arch = %v, want %v", got, expectedArch)
	}
	if got := d.Get("model"); got != expectedModel {
		t.Errorf("resourceLibvirtNodeInfoRead() model = %v, want %v", got, expectedModel)
	}
	if got := d.Get("vendor"); got != expectedVendor {
		t.Errorf("resourceLibvirtNodeInfoRead() vendor = %v, want %v", got, expectedVendor)
	}
	if got := d.Get("sockets"); got != expectedSockets {
		t.Errorf("resourceLibvirtNodeInfoRead() sockets = %v, want %v", got, expectedSockets)
	}
	if got := d.Get("dies"); got != expectedDies {
		t.Errorf("resourceLibvirtNodeInfoRead() dies = %v, want %v", got, expectedDies)
	}
	if got := d.Get("cores"); got != expectedCores {
		t.Errorf("resourceLibvirtNodeInfoRead() cores = %v, want %v", got, expectedCores)
	}
	if got := d.Get("threads"); got != expectedThreads {
		t.Errorf("resourceLibvirtNodeInfoRead() threads = %v, want %v", got, expectedThreads)
	}
	if got := d.Get("live_migration_support"); got != expectedLiveMigrationSupport {
		t.Errorf("resourceLibvirtNodeInfoRead() live_migration_support = %v, want %v", got, expectedLiveMigrationSupport)
	}
	if got := d.Get("live_migration_transports"); !reflect.DeepEqual(got, expectedLiveMigrationTransports) {
		t.Errorf("resourceLibvirtNodeInfoRead() live_migration_transports = %v, want %v", got, expectedLiveMigrationTransports)
	}
	if got := d.Get("features"); !reflect.DeepEqual(got, expectedFeatures) {
		t.Errorf("resourceLibvirtNodeInfoRead() features = %v, want %v", got, expectedFeatures)
	}
	if got := d.Id(); got != expectedUUID {
		t.Errorf("resourceLibvirtNodeInfoRead() UUID = %v, want %v", got, expectedUUID)
	}
}
