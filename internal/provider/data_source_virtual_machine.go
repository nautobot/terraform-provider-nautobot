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
	_ datasource.DataSource              = &VirtualMachineDataSource{}
	_ datasource.DataSourceWithConfigure = &VirtualMachineDataSource{}
)

type VirtualMachineDataSource struct {
	client *APIClient
}

type virtualMachineDataSourceModel struct {
	Name              types.String `tfsdk:"name"`
	ID                types.String `tfsdk:"id"`
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

func NewVirtualMachineDataSource() datasource.DataSource {
	return &VirtualMachineDataSource{}
}

func (d *VirtualMachineDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_virtual_machine"
}

func (d *VirtualMachineDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		Description: "Retrieves information about a specific virtual machine in Nautobot.",
		Attributes: map[string]dsschema.Attribute{
			"name": dsschema.StringAttribute{
				Description: "The name of the virtual machine to retrieve.",
				Required:    true,
			},
			"id": dsschema.StringAttribute{
				Description: "The UUID of the virtual machine.",
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
	}
}

func (d *VirtualMachineDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*APIClient)
}

func (d *VirtualMachineDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data virtualMachineDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if d.client == nil {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"API client is not configured. This is a bug in the provider configuration.",
		)
		return
	}

	c := d.client.Client

	vmName := data.Name.ValueString()

	// Fetch VM by name
	rsp, httpResp, err := c.VirtualizationAPI.
		VirtualizationVirtualMachinesList(ctx).
		Name([]string{vmName}).
		Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to get virtual machine",
			httpErr(err, httpResp),
		)
		return
	}

	if len(rsp.Results) == 0 {
		resp.Diagnostics.AddError(
			"Virtual machine not found",
			"No virtual machine found with name "+vmName,
		)
		return
	}

	vm := rsp.Results[0]

	if vm.Id == nil || *vm.Id == "" {
		resp.Diagnostics.AddError(
			"Invalid virtual machine data",
			"Virtual machine "+vmName+" returned no id",
		)
		return
	}
	vmID := *vm.Id
	data.ID = types.StringValue(vmID)

	data.Name = types.StringValue(vm.Name)

	if vm.Vcpus.IsSet() && vm.Vcpus.Get() != nil {
		data.Vcpus = types.Int64Value(int64(*vm.Vcpus.Get()))
	} else {
		data.Vcpus = types.Int64Value(0)
	}
	if vm.Memory.IsSet() && vm.Memory.Get() != nil {
		data.Memory = types.Int64Value(int64(*vm.Memory.Get()))
	} else {
		data.Memory = types.Int64Value(0)
	}
	if vm.Disk.IsSet() && vm.Disk.Get() != nil {
		data.Disk = types.Int64Value(int64(*vm.Disk.Get()))
	} else {
		data.Disk = types.Int64Value(0)
	}

	data.Comments = types.StringValue(derefStr(vm.Comments))
	data.Created = nullableTimeStr(vm.Created)
	data.LastUpdated = nullableTimeStr(vm.LastUpdated)

	clusterID := ""
	if vm.Cluster.Id != nil && vm.Cluster.Id.String != nil {
		clusterID = *vm.Cluster.Id.String
	}
	data.ClusterID = types.StringValue(clusterID)

	statusName := ""
	if vm.Status.Id != nil && vm.Status.Id.String != nil {
		statusID := *vm.Status.Id.String
		if statusID != "" {
			if n, err := getStatusName(ctx, c, statusID); err == nil {
				statusName = n
			}
		}
	}
	data.Status = types.StringValue(statusName)

	data.TenantID = nullableFKStr(vm.Tenant)
	data.PlatformID = nullableFKStr(vm.Platform)
	data.RoleID = nullableFKStr(vm.Role)
	data.SoftwareVersionID = nullableSoftwareVersionStr(vm.SoftwareVersion)

	primaryIPv4ID := ""
	if vm.PrimaryIp4.IsSet() {
		if ip4 := vm.PrimaryIp4.Get(); ip4 != nil && ip4.Id != nil && ip4.Id.String != nil {
			primaryIPv4ID = *ip4.Id.String
		}
	}
	data.PrimaryIP4ID = types.StringValue(primaryIPv4ID)

	primaryIPv6ID := ""
	if vm.PrimaryIp6.IsSet() {
		if ip6 := vm.PrimaryIp6.Get(); ip6 != nil && ip6.Id != nil && ip6.Id.String != nil {
			primaryIPv6ID = *ip6.Id.String
		}
	}
	data.PrimaryIP6ID = types.StringValue(primaryIPv6ID)

	if len(vm.Tags) > 0 {
		tagVals := make([]attr.Value, 0, len(vm.Tags))
		for _, tag := range vm.Tags {
			if tag.Id != nil && tag.Id.String != nil {
				tagVals = append(tagVals, types.StringValue(*tag.Id.String))
			}
		}
		data.TagsIDs = types.ListValueMust(types.StringType, tagVals)
	} else {
		data.TagsIDs = types.ListValueMust(types.StringType, []attr.Value{})
	}

	tflog.Debug(ctx, "read virtual machine", map[string]any{"id": vmID, "name": vmName})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
