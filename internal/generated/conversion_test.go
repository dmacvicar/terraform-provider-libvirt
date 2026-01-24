package generated

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"libvirt.org/go/libvirtxml"
)

// TestDomainTimerCatchUpRoundtrip tests conversion of a simple struct with no nesting
func TestDomainTimerCatchUpRoundtrip(t *testing.T) {
	ctx := context.Background()

	// Create original XML struct
	original := &libvirtxml.DomainTimerCatchUp{
		Threshold: 123,
		Slew:      120,
		Limit:     10000,
	}

	// Convert XML -> Model
	model, err := DomainTimerCatchUpFromXML(ctx, original, nil)
	if err != nil {
		t.Fatalf("FromXML failed: %v", err)
	}

	// Verify model fields
	if model.Threshold.IsNull() || model.Threshold.ValueInt64() != 123 {
		t.Errorf("Expected Threshold=123, got %v", model.Threshold)
	}
	if model.Slew.IsNull() || model.Slew.ValueInt64() != 120 {
		t.Errorf("Expected Slew=120, got %v", model.Slew)
	}
	if model.Limit.IsNull() || model.Limit.ValueInt64() != 10000 {
		t.Errorf("Expected Limit=10000, got %v", model.Limit)
	}

	// Convert Model -> XML
	converted, err := DomainTimerCatchUpToXML(ctx, model)
	if err != nil {
		t.Fatalf("ToXML failed: %v", err)
	}

	// Verify roundtrip
	if converted.Threshold != original.Threshold {
		t.Errorf("Roundtrip failed: Threshold %d != %d", converted.Threshold, original.Threshold)
	}
	if converted.Slew != original.Slew {
		t.Errorf("Roundtrip failed: Slew %d != %d", converted.Slew, original.Slew)
	}
	if converted.Limit != original.Limit {
		t.Errorf("Roundtrip failed: Limit %d != %d", converted.Limit, original.Limit)
	}
}

// TestDomainTimerNestedObjectRoundtrip tests conversion with nested objects
func TestDomainTimerNestedObjectRoundtrip(t *testing.T) {
	ctx := context.Background()

	// Create original XML struct with nested object
	original := &libvirtxml.DomainTimer{
		Name:       "rtc",
		Track:      "wall",
		TickPolicy: "catchup",
		CatchUp: &libvirtxml.DomainTimerCatchUp{
			Threshold: 123,
			Slew:      120,
			Limit:     10000,
		},
	}

	// Convert XML -> Model
	model, err := DomainTimerFromXML(ctx, original, nil)
	if err != nil {
		t.Fatalf("FromXML failed: %v", err)
	}

	// Verify simple fields
	if model.Name.ValueString() != "rtc" {
		t.Errorf("Expected Name=rtc, got %v", model.Name.ValueString())
	}

	// Verify nested object
	if model.CatchUp.IsNull() {
		t.Fatal("CatchUp should not be null")
	}

	// Convert Model -> XML
	converted, err := DomainTimerToXML(ctx, model)
	if err != nil {
		t.Fatalf("ToXML failed: %v", err)
	}

	// Verify roundtrip
	if converted.Name != original.Name {
		t.Errorf("Roundtrip failed: Name %s != %s", converted.Name, original.Name)
	}
	if converted.Track != original.Track {
		t.Errorf("Roundtrip failed: Track %s != %s", converted.Track, original.Track)
	}
	if converted.TickPolicy != original.TickPolicy {
		t.Errorf("Roundtrip failed: TickPolicy %s != %s", converted.TickPolicy, original.TickPolicy)
	}

	// Verify nested object roundtrip
	if converted.CatchUp == nil {
		t.Fatal("CatchUp should not be nil after roundtrip")
	}
	if converted.CatchUp.Threshold != original.CatchUp.Threshold {
		t.Errorf("Roundtrip failed: CatchUp.Threshold %d != %d", converted.CatchUp.Threshold, original.CatchUp.Threshold)
	}
}

// TestPreserveUserIntent tests that optional fields are only populated when user specified them
func TestPreserveUserIntent(t *testing.T) {
	ctx := context.Background()

	// Create XML with optional field set
	xmlWithOptional := &libvirtxml.DomainTimer{
		Name:       "rtc",
		Track:      "wall",
		TickPolicy: "catchup",
	}

	// Create plan with Track specified
	planWithTrack := &DomainTimerModel{
		Name:  types.StringValue("rtc"),
		Track: types.StringValue("wall"),
	}

	// Convert with plan - should preserve Track
	model1, err := DomainTimerFromXML(ctx, xmlWithOptional, planWithTrack)
	if err != nil {
		t.Fatalf("FromXML failed: %v", err)
	}
	if model1.Track.IsNull() {
		t.Error("Track should be preserved when in plan")
	}

	// Convert with plan that doesn't specify Track - should NOT populate Track
	planWithoutTrack := &DomainTimerModel{
		Name:  types.StringValue("rtc"),
		Track: types.StringNull(), // User didn't specify Track
	}
	model2, err := DomainTimerFromXML(ctx, xmlWithOptional, planWithoutTrack)
	if err != nil {
		t.Fatalf("FromXML failed: %v", err)
	}
	if !model2.Track.IsNull() {
		t.Error("Track should be null when not in plan")
	}

	// Convert with nil plan (import/datasource) - SHOULD populate all fields
	model3, err := DomainTimerFromXML(ctx, xmlWithOptional, nil)
	if err != nil {
		t.Fatalf("FromXML failed: %v", err)
	}
	if model3.Track.IsNull() {
		t.Error("Track should be populated when plan is nil (import/datasource mode)")
	}
	if model3.Track.ValueString() != "wall" {
		t.Errorf("Expected Track=wall, got %s", model3.Track.ValueString())
	}
}

func TestDomainAddressPCIZeroValues(t *testing.T) {
	ctx := context.Background()

	// Ensure zero-valued PCI address components are preserved when present in XML.
	domain := uint(0)
	bus := uint(3)
	slot := uint(0)
	function := uint(0)

	xml := &libvirtxml.DomainAddressPCI{
		Domain:   &domain,
		Bus:      &bus,
		Slot:     &slot,
		Function: &function,
	}

	plan := &DomainAddressPCIModel{
		Domain:   types.Int64Value(0),
		Bus:      types.Int64Value(3),
		Slot:     types.Int64Value(0),
		Function: types.Int64Value(0),
	}

	model, err := DomainAddressPCIFromXML(ctx, xml, plan)
	if err != nil {
		t.Fatalf("FromXML failed: %v", err)
	}

	if model.Domain.IsNull() || model.Domain.ValueInt64() != 0 {
		t.Errorf("Expected Domain=0, got %v", model.Domain)
	}
	if model.Bus.IsNull() || model.Bus.ValueInt64() != 3 {
		t.Errorf("Expected Bus=3, got %v", model.Bus)
	}
	if model.Slot.IsNull() || model.Slot.ValueInt64() != 0 {
		t.Errorf("Expected Slot=0, got %v", model.Slot)
	}
	if model.Function.IsNull() || model.Function.ValueInt64() != 0 {
		t.Errorf("Expected Function=0, got %v", model.Function)
	}
}

// TestNullHandling tests that null values are handled correctly
func TestNullHandling(t *testing.T) {
	ctx := context.Background()

	// Test null XML
	model, err := DomainTimerFromXML(ctx, nil, nil)
	if err != nil {
		t.Fatalf("FromXML with nil XML failed: %v", err)
	}
	if model != nil {
		t.Error("FromXML(nil) should return nil model")
	}

	// Test null Model
	xml, err := DomainTimerToXML(ctx, nil)
	if err != nil {
		t.Fatalf("ToXML with nil model failed: %v", err)
	}
	if xml != nil {
		t.Error("ToXML(nil) should return nil XML")
	}

	// Test nested null
	original := &libvirtxml.DomainTimer{
		Name:    "rtc",
		CatchUp: nil, // No nested object
	}

	model2, err := DomainTimerFromXML(ctx, original, nil)
	if err != nil {
		t.Fatalf("FromXML failed: %v", err)
	}
	if !model2.CatchUp.IsNull() {
		t.Error("CatchUp should be null when not present in XML")
	}

	// Convert back - should not have CatchUp
	converted, err := DomainTimerToXML(ctx, model2)
	if err != nil {
		t.Fatalf("ToXML failed: %v", err)
	}
	if converted.CatchUp != nil {
		t.Error("CatchUp should be nil when null in model")
	}
}

// TestPointerFields tests handling of pointer fields
func TestPointerFields(t *testing.T) {
	ctx := context.Background()

	// Many libvirtxml fields use pointers for optional values
	// Test that they convert correctly

	// Create XML with pointer field
	mode := "auto"
	original := &libvirtxml.DomainTimer{
		Name: "rtc",
		Mode: mode, // string field (not pointer in this case, but test the pattern)
	}

	model, err := DomainTimerFromXML(ctx, original, nil)
	if err != nil {
		t.Fatalf("FromXML failed: %v", err)
	}

	// Convert back
	converted, err := DomainTimerToXML(ctx, model)
	if err != nil {
		t.Fatalf("ToXML failed: %v", err)
	}

	// Verify pointer field roundtrip
	if converted.Mode != original.Mode {
		t.Errorf("Roundtrip failed: Mode %s != %s", converted.Mode, original.Mode)
	}
}

// TestNestedListRoundtrip tests conversion of lists of nested objects
func TestNestedListRoundtrip(t *testing.T) {
	ctx := context.Background()

	// Create original XML with a list of nested objects
	original := &libvirtxml.NetworkDNS{
		Enable: "yes",
		Forwarders: []libvirtxml.NetworkDNSForwarder{
			{
				Domain: "example.com",
				Addr:   "8.8.8.8",
			},
			{
				Domain: "test.com",
				Addr:   "1.1.1.1",
			},
		},
	}

	// Convert XML -> Model
	model, err := NetworkDNSFromXML(ctx, original, nil)
	if err != nil {
		t.Fatalf("FromXML failed: %v", err)
	}

	// Verify list is not null
	if model.Forwarders.IsNull() {
		t.Fatal("Forwarders list should not be null")
	}

	// Check list length
	var forwarders []NetworkDNSForwarderModel
	diags := model.Forwarders.ElementsAs(ctx, &forwarders, false)
	if diags.HasError() {
		t.Fatalf("ElementsAs failed: %v", diags)
	}
	if len(forwarders) != 2 {
		t.Fatalf("Expected 2 forwarders, got %d", len(forwarders))
	}

	// Verify first element
	if forwarders[0].Domain.ValueString() != "example.com" {
		t.Errorf("Expected Domain=example.com, got %s", forwarders[0].Domain.ValueString())
	}
	if forwarders[0].Addr.ValueString() != "8.8.8.8" {
		t.Errorf("Expected Addr=8.8.8.8, got %s", forwarders[0].Addr.ValueString())
	}

	// Convert Model -> XML
	converted, err := NetworkDNSToXML(ctx, model)
	if err != nil {
		t.Fatalf("ToXML failed: %v", err)
	}

	// Verify roundtrip
	if len(converted.Forwarders) != len(original.Forwarders) {
		t.Fatalf("Roundtrip failed: list length %d != %d", len(converted.Forwarders), len(original.Forwarders))
	}
	if converted.Forwarders[0].Domain != original.Forwarders[0].Domain {
		t.Errorf("Roundtrip failed: Domain %s != %s", converted.Forwarders[0].Domain, original.Forwarders[0].Domain)
	}
	if converted.Forwarders[0].Addr != original.Forwarders[0].Addr {
		t.Errorf("Roundtrip failed: Addr %s != %s", converted.Forwarders[0].Addr, original.Forwarders[0].Addr)
	}
}

// TestEmptyListRoundtrip tests that empty lists are handled correctly
func TestEmptyListRoundtrip(t *testing.T) {
	ctx := context.Background()

	// Create XML with empty list
	original := &libvirtxml.NetworkDNS{
		Enable:     "yes",
		Forwarders: []libvirtxml.NetworkDNSForwarder{},
	}

	// Convert XML -> Model
	model, err := NetworkDNSFromXML(ctx, original, nil)
	if err != nil {
		t.Fatalf("FromXML failed: %v", err)
	}

	// Empty list should result in null list (no elements)
	if !model.Forwarders.IsNull() {
		t.Error("Empty list should be null")
	}

	// Convert Model -> XML
	converted, err := NetworkDNSToXML(ctx, model)
	if err != nil {
		t.Fatalf("ToXML failed: %v", err)
	}

	// Should have no forwarders (or empty slice)
	if len(converted.Forwarders) != 0 {
		t.Errorf("Expected empty forwarders list, got %d items", len(converted.Forwarders))
	}
}

// TestNilListHandling tests that nil lists are handled correctly
func TestNilListHandling(t *testing.T) {
	ctx := context.Background()

	// Create XML with nil list
	original := &libvirtxml.NetworkDNS{
		Enable:     "yes",
		Forwarders: nil,
	}

	// Convert XML -> Model
	model, err := NetworkDNSFromXML(ctx, original, nil)
	if err != nil {
		t.Fatalf("FromXML failed: %v", err)
	}

	// Nil list should result in null list
	if !model.Forwarders.IsNull() {
		t.Error("Nil list should be null")
	}

	// Convert Model -> XML
	converted, err := NetworkDNSToXML(ctx, model)
	if err != nil {
		t.Fatalf("ToXML failed: %v", err)
	}

	// Should have no forwarders
	if len(converted.Forwarders) != 0 {
		t.Errorf("Expected no forwarders, got %d items", len(converted.Forwarders))
	}
}

// NOTE: TestRealLibvirtXML is skipped because Domain has circular references
// (DomainDiskSource → DomainDiskDataStore → DomainDiskSource) which cause stack overflow
// in AttributeTypes() generation. This is a known limitation that needs to be addressed
// in the generator. For now, the simpler structures (Timer, Clock, NetworkDNS, etc.)
// demonstrate that the generated code works correctly for non-circular structures.

// TestComplexNestedStructure tests a complex nested structure
func TestComplexNestedStructure(t *testing.T) {
	ctx := context.Background()

	// Create a complex structure with multiple levels of nesting
	original := &libvirtxml.DomainClock{
		Offset: "utc",
		Timer: []libvirtxml.DomainTimer{
			{
				Name:       "rtc",
				TickPolicy: "catchup",
				CatchUp: &libvirtxml.DomainTimerCatchUp{
					Threshold: 123,
					Slew:      120,
					Limit:     10000,
				},
			},
			{
				Name:       "pit",
				TickPolicy: "delay",
			},
		},
	}

	// Convert to model
	model, err := DomainClockFromXML(ctx, original, nil)
	if err != nil {
		t.Fatalf("DomainClockFromXML failed: %v", err)
	}

	// Verify we have timers
	if model.Timer.IsNull() {
		t.Fatal("Timer list should not be null")
	}

	var timers []DomainTimerModel
	diags := model.Timer.ElementsAs(ctx, &timers, false)
	if diags.HasError() {
		t.Fatalf("ElementsAs failed: %v", diags)
	}
	if len(timers) != 2 {
		t.Fatalf("Expected 2 timers, got %d", len(timers))
	}

	// Verify first timer has nested CatchUp
	if timers[0].CatchUp.IsNull() {
		t.Error("First timer should have CatchUp")
	}

	// Verify second timer doesn't have CatchUp
	if !timers[1].CatchUp.IsNull() {
		t.Error("Second timer should not have CatchUp")
	}

	// Convert back to XML
	converted, err := DomainClockToXML(ctx, model)
	if err != nil {
		t.Fatalf("DomainClockToXML failed: %v", err)
	}

	// Verify roundtrip
	if len(converted.Timer) != len(original.Timer) {
		t.Fatalf("Roundtrip failed: timer count %d != %d", len(converted.Timer), len(original.Timer))
	}
	if converted.Timer[0].Name != original.Timer[0].Name {
		t.Errorf("Roundtrip failed: Timer[0].Name %s != %s", converted.Timer[0].Name, original.Timer[0].Name)
	}
	if converted.Timer[0].CatchUp == nil || original.Timer[0].CatchUp == nil {
		t.Error("Both should have CatchUp")
	} else {
		if converted.Timer[0].CatchUp.Threshold != original.Timer[0].CatchUp.Threshold {
			t.Errorf("Roundtrip failed: CatchUp.Threshold %d != %d",
				converted.Timer[0].CatchUp.Threshold, original.Timer[0].CatchUp.Threshold)
		}
	}
}

// TestMemoryValueWithUnitRoundtrip verifies flattened value/unit pairs stay consistent.
func TestMemoryValueWithUnitRoundtrip(t *testing.T) {
	ctx := context.Background()

	original := &libvirtxml.Domain{
		Type: "kvm",
		Name: "test-memory",
		Memory: &libvirtxml.DomainMemory{
			Value:    524288,
			Unit:     "KiB",
			DumpCore: "off",
		},
	}

	model, err := DomainFromXML(ctx, original, nil)
	if err != nil {
		t.Fatalf("DomainFromXML failed: %v", err)
	}

	if model.Memory.IsNull() || model.Memory.ValueInt64() != 524288 {
		t.Fatalf("expected flattened memory value, got %v", model.Memory)
	}
	if model.MemoryUnit.IsNull() || model.MemoryUnit.ValueString() != "KiB" {
		t.Fatalf("expected flattened memory unit, got %v", model.MemoryUnit)
	}
	if model.MemoryDumpCore.IsNull() || model.MemoryDumpCore.ValueString() != "off" {
		t.Fatalf("expected flattened memory_dump_core, got %v", model.MemoryDumpCore)
	}

	converted, err := DomainToXML(ctx, model)
	if err != nil {
		t.Fatalf("DomainToXML failed: %v", err)
	}
	if converted.Memory == nil {
		t.Fatalf("converted domain missing memory")
	}
	if converted.Memory.Value != 524288 || converted.Memory.Unit != "KiB" || converted.Memory.DumpCore != "off" {
		t.Fatalf("unexpected converted memory: %+v", converted.Memory)
	}
}

// TestBooleanPresenceRoundtrip verifies boolean-to-presence conversions.
func TestBooleanPresenceRoundtrip(t *testing.T) {
	ctx := context.Background()

	original := &libvirtxml.DomainRNGBackend{
		BuiltIn: &libvirtxml.DomainRNGBackendBuiltIn{},
	}

	model, err := DomainRNGBackendFromXML(ctx, original, nil)
	if err != nil {
		t.Fatalf("DomainRNGBackendFromXML failed: %v", err)
	}
	if model.BuiltIn.IsNull() || !model.BuiltIn.ValueBool() {
		t.Fatalf("expected BuiltIn to be true, got %v", model.BuiltIn)
	}

	converted, err := DomainRNGBackendToXML(ctx, model)
	if err != nil {
		t.Fatalf("DomainRNGBackendToXML failed: %v", err)
	}
	if converted.BuiltIn == nil {
		t.Fatal("BuiltIn should be present when bool true")
	}

	model.BuiltIn = types.BoolValue(false)
	convertedFalse, err := DomainRNGBackendToXML(ctx, model)
	if err != nil {
		t.Fatalf("DomainRNGBackendToXML(false) failed: %v", err)
	}
	if convertedFalse.BuiltIn != nil {
		t.Fatal("BuiltIn element should be omitted when bool false")
	}
}

// TestDomainDeviceListFromXMLPreservesExplicitEmptyList verifies that an explicitly
// empty list stays empty (not null) when reading from XML. This uses a concrete
// DomainDeviceList/Hostdevs structure to cover the abstract "empty list vs null" case.
func TestDomainDeviceListFromXMLPreservesExplicitEmptyList(t *testing.T) {
	ctx := context.Background()

	emptyHostdevs, diags := types.ListValueFrom(
		ctx,
		types.ObjectType{AttrTypes: DomainHostdevAttributeTypes()},
		[]DomainHostdevModel{},
	)
	if diags.HasError() {
		t.Fatalf("ListValueFrom failed: %s", diags.Errors()[0].Summary())
	}

	plan := &DomainDeviceListModel{
		Hostdevs: emptyHostdevs,
	}
	original := &libvirtxml.DomainDeviceList{}

	model, err := DomainDeviceListFromXML(ctx, original, plan)
	if err != nil {
		t.Fatalf("DomainDeviceListFromXML failed: %v", err)
	}
	if model.Hostdevs.IsNull() {
		t.Fatal("expected Hostdevs to be an empty list, got null")
	}

	var got []DomainHostdevModel
	if diags := model.Hostdevs.ElementsAs(ctx, &got, false); diags.HasError() {
		t.Fatalf("ElementsAs failed: %s", diags.Errors()[0].Summary())
	}
	if len(got) != 0 {
		t.Fatalf("expected zero hostdevs, got %d", len(got))
	}
}

func TestDomainDeviceListFromXMLPreservesDiskOrderByTargetDev(t *testing.T) {
	ctx := context.Background()

	makeTarget := func(dev string) types.Object {
		target := DomainDiskTargetModel{
			Dev:          types.StringValue(dev),
			Bus:          types.StringNull(),
			Tray:         types.StringNull(),
			Removable:    types.StringNull(),
			RotationRate: types.Int64Null(),
		}
		obj, diags := types.ObjectValueFrom(ctx, DomainDiskTargetAttributeTypes(), target)
		if diags.HasError() {
			t.Fatalf("target ObjectValueFrom failed: %s", diags.Errors()[0].Summary())
		}
		return obj
	}

	makeSourceVolume := func(pool, volume string) types.Object {
		vol := DomainDiskSourceVolumeModel{
			Pool:     types.StringValue(pool),
			Volume:   types.StringValue(volume),
			Mode:     types.StringNull(),
			SecLabel: types.ListNull(types.ObjectType{AttrTypes: DomainDeviceSecLabelAttributeTypes()}),
		}
		volObj, diags := types.ObjectValueFrom(ctx, DomainDiskSourceVolumeAttributeTypes(), vol)
		if diags.HasError() {
			t.Fatalf("volume ObjectValueFrom failed: %s", diags.Errors()[0].Summary())
		}

		source := DomainDiskSourceModel{
			File:          types.ObjectNull(DomainDiskSourceFileAttributeTypes()),
			Block:         types.ObjectNull(DomainDiskSourceBlockAttributeTypes()),
			Dir:           types.ObjectNull(DomainDiskSourceDirAttributeTypes()),
			Network:       types.ObjectNull(DomainDiskSourceNetworkAttributeTypes()),
			Volume:        volObj,
			NVME:          types.ObjectNull(DomainDiskSourceNVMEAttributeTypes()),
			VHostUser:     types.ObjectNull(DomainDiskSourceVHostUserAttributeTypes()),
			VHostVDPA:     types.ObjectNull(DomainDiskSourceVHostVDPAAttributeTypes()),
			StartupPolicy: types.StringNull(),
			Index:         types.Int64Null(),
			Encryption:    types.ObjectNull(DomainDiskEncryptionAttributeTypes()),
			Reservations:  types.ObjectNull(DomainDiskReservationsAttributeTypes()),
			Slices:        types.ObjectNull(DomainDiskSlicesAttributeTypes()),
			SSL:           types.ObjectNull(DomainDiskSourceSSLAttributeTypes()),
			Cookies:       types.ObjectNull(DomainDiskCookiesAttributeTypes()),
			Readahead:     types.ObjectNull(DomainDiskSourceReadaheadAttributeTypes()),
			Timeout:       types.ObjectNull(DomainDiskSourceTimeoutAttributeTypes()),
			DataStore:     types.ObjectNull(DomainDiskDataStoreAttributeTypes()),
		}
		sourceObj, diags := types.ObjectValueFrom(ctx, DomainDiskSourceAttributeTypes(), source)
		if diags.HasError() {
			t.Fatalf("source ObjectValueFrom failed: %s", diags.Errors()[0].Summary())
		}
		return sourceObj
	}

	makeDisk := func(deviceType, dev, pool, volume string) DomainDiskModel {
		diskDevice := types.StringNull()
		if deviceType != "" {
			diskDevice = types.StringValue(deviceType)
		}
		return DomainDiskModel{
			Device:          diskDevice,
			RawIO:           types.StringNull(),
			SGIO:            types.StringNull(),
			Snapshot:        types.StringNull(),
			Model:           types.StringNull(),
			Driver:          types.ObjectNull(DomainDiskDriverAttributeTypes()),
			Auth:            types.ObjectNull(DomainDiskAuthAttributeTypes()),
			Source:          makeSourceVolume(pool, volume),
			BackingStore:    types.ObjectNull(DomainDiskBackingStoreAttributeTypes()),
			BackendDomain:   types.ObjectNull(DomainBackendDomainAttributeTypes()),
			Geometry:        types.ObjectNull(DomainDiskGeometryAttributeTypes()),
			BlockIO:         types.ObjectNull(DomainDiskBlockIOAttributeTypes()),
			Mirror:          types.ObjectNull(DomainDiskMirrorAttributeTypes()),
			Target:          makeTarget(dev),
			IOTune:          types.ObjectNull(DomainDiskIOTuneAttributeTypes()),
			ThrottleFilters: types.ObjectNull(ThrottleFiltersAttributeTypes()),
			ReadOnly:        types.BoolNull(),
			Shareable:       types.BoolNull(),
			Transient:       types.ObjectNull(DomainDiskTransientAttributeTypes()),
			Serial:          types.StringNull(),
			WWN:             types.StringNull(),
			Vendor:          types.StringNull(),
			Product:         types.StringNull(),
			Encryption:      types.ObjectNull(DomainDiskEncryptionAttributeTypes()),
			Boot:            types.ObjectNull(DomainDeviceBootAttributeTypes()),
			ACPI:            types.ObjectNull(DomainDeviceACPIAttributeTypes()),
			Alias:           types.ObjectNull(DomainAliasAttributeTypes()),
			Address:         types.ObjectNull(DomainAddressAttributeTypes()),
		}
	}

	planDisks := []DomainDiskModel{
		makeDisk("disk", "vda", "ssd", "vm-01_0.qcow2"),
		makeDisk("cdrom", "sda", "ssd", "vm-01-cloudinit.iso"),
		makeDisk("disk", "vdb", "sata", "vm-01_1.qcow2"),
	}

	planList, diags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: DomainDiskAttributeTypes()}, planDisks)
	if diags.HasError() {
		t.Fatalf("plan ListValueFrom failed: %s", diags.Errors()[0].Summary())
	}

	plan := &DomainDeviceListModel{
		Disks: planList,
	}

	original := &libvirtxml.DomainDeviceList{
		Disks: []libvirtxml.DomainDisk{
			{
				Device: "disk",
				Target: &libvirtxml.DomainDiskTarget{Dev: "vda"},
				Source: &libvirtxml.DomainDiskSource{Volume: &libvirtxml.DomainDiskSourceVolume{Pool: "ssd", Volume: "vm-01_0.qcow2"}},
			},
			{
				Device: "disk",
				Target: &libvirtxml.DomainDiskTarget{Dev: "vdb"},
				Source: &libvirtxml.DomainDiskSource{Volume: &libvirtxml.DomainDiskSourceVolume{Pool: "sata", Volume: "vm-01_1.qcow2"}},
			},
			{
				Device: "cdrom",
				Target: &libvirtxml.DomainDiskTarget{Dev: "sda"},
				Source: &libvirtxml.DomainDiskSource{Volume: &libvirtxml.DomainDiskSourceVolume{Pool: "ssd", Volume: "vm-01-cloudinit.iso"}},
			},
		},
	}

	model, err := DomainDeviceListFromXML(ctx, original, plan)
	if err != nil {
		t.Fatalf("DomainDeviceListFromXML failed: %v", err)
	}

	var got []DomainDiskModel
	if diags := model.Disks.ElementsAs(ctx, &got, false); diags.HasError() {
		t.Fatalf("ElementsAs failed: %s", diags.Errors()[0].Summary())
	}

	gotDevs := make([]string, 0, len(got))
	for _, disk := range got {
		var target DomainDiskTargetModel
		if diags := disk.Target.As(ctx, &target, basetypes.ObjectAsOptions{}); diags.HasError() {
			t.Fatalf("disk target As failed: %s", diags.Errors()[0].Summary())
		}
		gotDevs = append(gotDevs, target.Dev.ValueString())
	}
	if len(gotDevs) != 3 || gotDevs[0] != "vda" || gotDevs[1] != "sda" || gotDevs[2] != "vdb" {
		t.Fatalf("unexpected disk order: %v", gotDevs)
	}

	var secondSource DomainDiskSourceModel
	if diags := got[1].Source.As(ctx, &secondSource, basetypes.ObjectAsOptions{}); diags.HasError() {
		t.Fatalf("disk source As failed: %s", diags.Errors()[0].Summary())
	}
	var secondVolume DomainDiskSourceVolumeModel
	if diags := secondSource.Volume.As(ctx, &secondVolume, basetypes.ObjectAsOptions{}); diags.HasError() {
		t.Fatalf("disk volume As failed: %s", diags.Errors()[0].Summary())
	}
	if secondVolume.Pool.ValueString() != "ssd" || secondVolume.Volume.ValueString() != "vm-01-cloudinit.iso" {
		t.Fatalf("unexpected second disk volume: %s/%s", secondVolume.Pool.ValueString(), secondVolume.Volume.ValueString())
	}
}
