package provider

import (
	"context"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	nb "github.com/nautobot/go-nautobot/v2"
)

func dataSourceVirtualMachines() *schema.Resource {
	return &schema.Resource{
		Description: "Retrieves information about virtual machines in Nautobot.",

		ReadContext: dataSourceVirtualMachinesRead,

		Schema: map[string]*schema.Schema{
			"virtual_machines": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Description: "The UUID of the virtual machine.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"name": {
							Description: "The name of the virtual machine.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"cluster_id": {
							Description: "The ID of the cluster associated with the virtual machine.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"status": {
							Description: "The name of the status of the virtual machine.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"tenant_id": {
							Description: "The ID of the tenant associated with the virtual machine.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"platform_id": {
							Description: "The ID of the platform associated with the virtual machine.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"role_id": {
							Description: "The ID of the role associated with the virtual machine.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"primary_ip4_id": {
							Description: "The ID of the primary IPv4 address.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"primary_ip6_id": {
							Description: "The ID of the primary IPv6 address.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"vcpus": {
							Description: "The number of virtual CPUs.",
							Type:        schema.TypeInt,
							Computed:    true,
						},
						"memory": {
							Description: "The amount of memory in MB.",
							Type:        schema.TypeInt,
							Computed:    true,
						},
						"disk": {
							Description: "The disk size in GB.",
							Type:        schema.TypeInt,
							Computed:    true,
						},
						"comments": {
							Description: "Comments or notes about the virtual machine.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"tags_ids": {
							Description: "The IDs of the tags associated with the virtual machine.",
							Type:        schema.TypeList,
							Computed:    true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"created": {
							Description: "The creation date of the virtual machine.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"last_updated": {
							Description: "The last update date of the virtual machine.",
							Type:        schema.TypeString,
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func dataSourceVirtualMachinesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c := meta.(*apiClient).Client
	s := meta.(*apiClient).Server
	t := meta.(*apiClient).Token.token

	// Auth context
	auth := context.WithValue(
		ctx,
		nb.ContextAPIKeys,
		map[string]nb.APIKey{
			"tokenAuth": {
				Key:    t,
				Prefix: "Token",
			},
		},
	)

	// Fetch virtual machines list
	rsp, _, err := c.VirtualizationAPI.VirtualizationVirtualMachinesList(auth).Execute()
	if err != nil {
		return diag.Errorf("failed to get virtual machines list from %s: %s", s, err.Error())
	}

	results := rsp.Results
	list := make([]map[string]interface{}, 0)

	for _, vm := range results {
		createdStr := ""
		if vm.Created.IsSet() && vm.Created.Get() != nil {
			createdStr = vm.Created.Get().Format(time.RFC3339)
		}

		lastUpdatedStr := ""
		if vm.LastUpdated.IsSet() && vm.LastUpdated.Get() != nil {
			lastUpdatedStr = vm.LastUpdated.Get().Format(time.RFC3339)
		}

		itemMap := map[string]interface{}{
			"id":           vm.Id,
			"name":         vm.Name,
			"vcpus":        vm.Vcpus.Get(),
			"memory":       vm.Memory.Get(),
			"disk":         vm.Disk.Get(),
			"comments":     vm.Comments,
			"created":      createdStr,
			"last_updated": lastUpdatedStr,
			"tags_ids":     vm.Tags,
		}

		// Extract cluster_id, status, and other fields
		if vm.Cluster.Id != nil && vm.Cluster.Id.String != nil {
			itemMap["cluster_id"] = *vm.Cluster.Id.String
		}

		if vm.Status.Id != nil && vm.Status.Id.String != nil {
			statusID := *vm.Status.Id.String
			statusName, err := getStatusName(ctx, c, t, statusID)
			if err != nil {
				return diag.Errorf("failed to get status name for ID %s: %s", statusID, err.Error())
			}
			itemMap["status"] = statusName
		}

		// Handle nullable fields (tenant, platform, role, etc.)
		if vm.Tenant.IsSet() {
			tenant := vm.Tenant.Get()
			if tenant != nil && tenant.Id != nil {
				itemMap["tenant_id"] = *tenant.Id.String
			}
		}

		if vm.Platform.IsSet() {
			platform := vm.Platform.Get()
			if platform != nil && platform.Id != nil {
				itemMap["platform_id"] = *platform.Id.String
			}
		}

		if vm.Role.IsSet() {
			role := vm.Role.Get()
			if role != nil && role.Id != nil {
				itemMap["role_id"] = *role.Id.String
			}
		}

		if vm.PrimaryIp4.IsSet() {
			primaryIp4 := vm.PrimaryIp4.Get()
			if primaryIp4 != nil && primaryIp4.Id != nil {
				itemMap["primary_ip4_id"] = *primaryIp4.Id.String
			}
		}

		if vm.PrimaryIp6.IsSet() {
			primaryIp6 := vm.PrimaryIp6.Get()
			if primaryIp6 != nil && primaryIp6.Id != nil {
				itemMap["primary_ip6_id"] = *primaryIp6.Id.String
			}
		}

		list = append(list, itemMap)
	}

	if err := d.Set("virtual_machines", list); err != nil {
		return diag.FromErr(err)
	}

	// Set ID for the data source
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}
