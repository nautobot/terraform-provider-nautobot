package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ datasource.DataSource              = &VirtualMachinesDataSource{}
	_ datasource.DataSourceWithConfigure = &VirtualMachinesDataSource{}
)

type VirtualMachinesDataSource struct {
	client *APIClient
}

type virtualMachineItemModel struct {
	ID                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	ClusterID         types.String `tfsdk:"cluster_id"`
	Status            types.String `tfsdk:"status"`
	TenantID          types.String `tfsdk:"tenant_id"`
	PlatformID        types.String `tfsdk:"platform_id"`
	RoleID            types.String `tfsdk:"role_id"`
	SoftwareVersionID types.String `tfsdk:"software_version_id"`
	PrimaryIP4ID      types.String `tfsdk:"primary_ip4_id"`
	PrimaryIP6ID      types.String `tfsdk:"primary_ip6_id"`
	Vcpus             types.Int64  `tfsdk:"vcpus"`
	Memory            types.Int64  `tfsdk:"memory"`
	Disk              types.Int64  `tfsdk:"disk"`
	Comments          types.String `tfsdk:"comments"`
	TagsIDs           types.List   `tfsdk:"tags_ids"`
	Created           types.String `tfsdk:"created"`
	LastUpdated       types.String `tfsdk:"last_updated"`
}

type virtualMachinesDataSourceModel struct {
	VirtualMachines []virtualMachineItemModel `tfsdk:"virtual_machines"`
}

func NewVirtualMachinesDataSource() datasource.DataSource {
	return &VirtualMachinesDataSource{}
}

func (d *VirtualMachinesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_virtual_machines"
}

func (d *VirtualMachinesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		Description: "Retrieves information about all virtual machines in Nautobot.",
		Attributes: map[string]dsschema.Attribute{
			"virtual_machines": dsschema.ListNestedAttribute{
				Description: "List of virtual machines.",
				Computed:    true,
				NestedObject: dsschema.NestedAttributeObject{
					Attributes: map[string]dsschema.Attribute{
						"id": dsschema.StringAttribute{
							Description: "The UUID of the virtual machine.",
							Computed:    true,
						},
						"name": dsschema.StringAttribute{
							Description: "The name of the virtual machine.",
							Computed:    true,
						},
						"cluster_id": dsschema.StringAttribute{
							Description: "The ID of the cluster associated with the virtual machine.",
							Computed:    true,
						},
						"status": dsschema.StringAttribute{
							Description: "The name of the status of the virtual machine.",
							Computed:    true,
						},
						"tenant_id": dsschema.StringAttribute{
							Description: "The ID of the tenant associated with the virtual machine.",
							Computed:    true,
						},
						"platform_id": dsschema.StringAttribute{
							Description: "The ID of the platform associated with the virtual machine.",
							Computed:    true,
						},
						"role_id": dsschema.StringAttribute{
							Description: "The ID of the role associated with the virtual machine.",
							Computed:    true,
						},
						"software_version_id": dsschema.StringAttribute{
							Description: "The ID of the software version installed on the virtual machine.",
							Computed:    true,
						},
						"primary_ip4_id": dsschema.StringAttribute{
							Description: "The ID of the primary IPv4 address.",
							Computed:    true,
						},
						"primary_ip6_id": dsschema.StringAttribute{
							Description: "The ID of the primary IPv6 address.",
							Computed:    true,
						},
						"vcpus": dsschema.Int64Attribute{
							Description: "The number of virtual CPUs.",
							Computed:    true,
						},
						"memory": dsschema.Int64Attribute{
							Description: "The amount of memory in MB.",
							Computed:    true,
						},
						"disk": dsschema.Int64Attribute{
							Description: "The disk size in GB.",
							Computed:    true,
						},
						"comments": dsschema.StringAttribute{
							Description: "Comments or notes about the virtual machine.",
							Computed:    true,
						},
						"tags_ids": dsschema.ListAttribute{
							Description: "The IDs of the tags associated with the virtual machine.",
							Computed:    true,
							ElementType: types.StringType,
						},
						"created": dsschema.StringAttribute{
							Description: "The creation date of the virtual machine (RFC3339).",
							Computed:    true,
						},
						"last_updated": dsschema.StringAttribute{
							Description: "The last update date of the virtual machine (RFC3339).",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *VirtualMachinesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*APIClient)
}

func (d *VirtualMachinesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state virtualMachinesDataSourceModel

	if d.client == nil {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"API client is not configured. This is a bug in the provider configuration.",
		)
		return
	}

	c := d.client.Client

	const pageLimit int32 = 200
	var offset int32 = 0

	state.VirtualMachines = make([]virtualMachineItemModel, 0)

	for {
		rsp, httpResp, err := c.VirtualizationAPI.
			VirtualizationVirtualMachinesList(ctx).
			Limit(pageLimit).
			Offset(offset).
			Sort("name").
			Execute()
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to get virtual machines list",
				httpErr(err, httpResp),
			)
			return
		}

		results := rsp.Results
		if len(results) == 0 {
			break
		}

		for _, vm := range results {
			var item virtualMachineItemModel

			if vm.Id == nil || *vm.Id == "" {
				resp.Diagnostics.AddError(
					"Invalid virtual machine data",
					"Virtual machines list returned an item with no id (name: "+vm.Name+")",
				)
				return
			}
			item.ID = types.StringValue(*vm.Id)

			item.Created = nullableTimeStr(vm.Created)
			item.LastUpdated = nullableTimeStr(vm.LastUpdated)

			item.Name = types.StringValue(vm.Name)

			vcpusVal := int64(0)
			if vm.Vcpus.IsSet() && vm.Vcpus.Get() != nil {
				vcpusVal = int64(*vm.Vcpus.Get())
			}
			item.Vcpus = types.Int64Value(vcpusVal)

			memoryVal := int64(0)
			if vm.Memory.IsSet() && vm.Memory.Get() != nil {
				memoryVal = int64(*vm.Memory.Get())
			}
			item.Memory = types.Int64Value(memoryVal)

			diskVal := int64(0)
			if vm.Disk.IsSet() && vm.Disk.Get() != nil {
				diskVal = int64(*vm.Disk.Get())
			}
			item.Disk = types.Int64Value(diskVal)

			item.Comments = types.StringValue(derefStr(vm.Comments))

			clusterID := ""
			if vm.Cluster.Id != nil && vm.Cluster.Id.String != nil {
				clusterID = *vm.Cluster.Id.String
			}
			item.ClusterID = types.StringValue(clusterID)

			statusName := ""
			if vm.Status.Id != nil && vm.Status.Id.String != nil {
				statusID := *vm.Status.Id.String
				if statusID != "" {
					if n, err := getStatusName(ctx, c, statusID); err == nil {
						statusName = n
					}
				}
			}
			item.Status = types.StringValue(statusName)

			item.TenantID = nullableFKStr(vm.Tenant)
			item.PlatformID = nullableFKStr(vm.Platform)
			item.RoleID = nullableFKStr(vm.Role)
			item.SoftwareVersionID = nullableSoftwareVersionStr(vm.SoftwareVersion)

			primaryIPv4ID := ""
			if vm.PrimaryIp4.IsSet() {
				if ip4 := vm.PrimaryIp4.Get(); ip4 != nil && ip4.Id != nil && ip4.Id.String != nil {
					primaryIPv4ID = *ip4.Id.String
				}
			}
			item.PrimaryIP4ID = types.StringValue(primaryIPv4ID)

			primaryIPv6ID := ""
			if vm.PrimaryIp6.IsSet() {
				if ip6 := vm.PrimaryIp6.Get(); ip6 != nil && ip6.Id != nil && ip6.Id.String != nil {
					primaryIPv6ID = *ip6.Id.String
				}
			}
			item.PrimaryIP6ID = types.StringValue(primaryIPv6ID)

			if len(vm.Tags) > 0 {
				tagVals := make([]attr.Value, 0, len(vm.Tags))
				for _, tag := range vm.Tags {
					if tag.Id != nil && tag.Id.String != nil {
						tagVals = append(tagVals, types.StringValue(*tag.Id.String))
					}
				}
				item.TagsIDs = types.ListValueMust(types.StringType, tagVals)
			} else {
				item.TagsIDs = types.ListValueMust(types.StringType, []attr.Value{})
			}

			state.VirtualMachines = append(state.VirtualMachines, item)
		}

		offset += int32(len(results))

		if !rsp.Next.IsSet() || rsp.Next.Get() == nil || *rsp.Next.Get() == "" {
			break
		}
	}

	tflog.Debug(ctx, "read virtual machines", map[string]any{"count": len(state.VirtualMachines)})

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
