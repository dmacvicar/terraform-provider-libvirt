package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/dmacvicar/terraform-provider-libvirt/v2/internal/libvirt"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &NodeInfoDataSource{}

func NewNodeInfoDataSource() datasource.DataSource {
	return &NodeInfoDataSource{}
}

type NodeInfoDataSource struct {
	client *libvirt.Client
}

type NodeInfoDataSourceModel struct {
	ID                  types.String `tfsdk:"id"`
	CPUModel            types.String `tfsdk:"cpu_model"`
	MemoryTotalKB       types.Int64  `tfsdk:"memory_total_kb"`
	CPUCoresTotal       types.Int64  `tfsdk:"cpu_cores_total"`
	NumaNodes           types.Int64  `tfsdk:"numa_nodes"`
	CPUSockets          types.Int64  `tfsdk:"cpu_sockets"`
	CPUCoresPerSocket   types.Int64  `tfsdk:"cpu_cores_per_socket"`
	CPUThreadsPerCore   types.Int64  `tfsdk:"cpu_threads_per_core"`
}

func (d *NodeInfoDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_node_info"
}

func (d *NodeInfoDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches information about the libvirt host node.\n\n" +
			"This data source provides details about the host system's hardware capabilities, " +
			"including CPU model, memory, and processor topology.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Internal identifier for this data source (hash of all values).",
			},
			"cpu_model": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "CPU model name (e.g., 'x86_64').",
			},
			"memory_total_kb": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Total memory available on the host in kilobytes.",
			},
			"cpu_cores_total": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Total number of logical CPU cores available on the host.",
			},
			"numa_nodes": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Number of NUMA nodes on the host.",
			},
			"cpu_sockets": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Number of CPU sockets on the host.",
			},
			"cpu_cores_per_socket": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Number of CPU cores per socket.",
			},
			"cpu_threads_per_core": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Number of threads per CPU core (e.g., 2 for hyper-threading).",
			},
		},
	}
}

func (d *NodeInfoDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*libvirt.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *libvirt.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *NodeInfoDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data NodeInfoDataSourceModel

	// Call NodeGetInfo API
	model, memory, cpus, _, nodes, sockets, cores, threads, err := d.client.Libvirt().NodeGetInfo()
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to get node information",
			fmt.Sprintf("Unable to retrieve host node information: %s", err),
		)
		return
	}

	// Convert model from [32]int8 to string
	cpuModel := int8ArrayToString(model)

	// Populate model
	data.CPUModel = types.StringValue(cpuModel)
	data.MemoryTotalKB = types.Int64Value(int64(memory))
	data.CPUCoresTotal = types.Int64Value(int64(cpus))
	data.NumaNodes = types.Int64Value(int64(nodes))
	data.CPUSockets = types.Int64Value(int64(sockets))
	data.CPUCoresPerSocket = types.Int64Value(int64(cores))
	data.CPUThreadsPerCore = types.Int64Value(int64(threads))

	// Generate ID as hash of all values
	idStr := fmt.Sprintf("%s-%d-%d-%d-%d-%d-%d",
		cpuModel, memory, cpus, nodes, sockets, cores, threads)
	data.ID = types.StringValue(strconv.Itoa(hashString(idStr)))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// int8ArrayToString converts a null-terminated [32]int8 array to a Go string.
// This is used to convert the CPU model name from libvirt's NodeGetInfo API.
func int8ArrayToString(arr [32]int8) string {
	// Find the null terminator and convert to bytes
	var bytes []byte
	for _, b := range arr {
		if b == 0 {
			break
		}
		bytes = append(bytes, byte(b))
	}
	return string(bytes)
}

// hashString returns a simple hash of the input string.
// This is used for generating stable IDs for data sources.
func hashString(s string) int {
	h := 0
	for _, c := range s {
		h = 31*h + int(c)
	}
	if h < 0 {
		h = -h
	}
	return h
}
