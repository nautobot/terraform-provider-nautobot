package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	nb "github.com/nautobot/go-nautobot/v2"
)

func resourcePrimaryIPAddressForVM() *schema.Resource {
	return &schema.Resource{
		Description: "This resource sets an IP address as the primary IPv4 or IPv6 for a virtual machine in Nautobot",

		CreateContext: resourcePrimaryIPAddressForVMCreate,
		ReadContext:   resourcePrimaryIPAddressForVMRead,
		UpdateContext: resourcePrimaryIPAddressForVMUpdate,
		DeleteContext: resourcePrimaryIPAddressForVMDelete,

		Schema: map[string]*schema.Schema{
			"virtual_machine_id": {
				Description: "ID of the virtual machine.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"primary_ip4_id": {
				Description: "ID of the primary IPv4 address.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"primary_ip6_id": {
				Description: "ID of the primary IPv6 address.",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}
}

func resourcePrimaryIPAddressForVMCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*apiClient).Client
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

	vmID := d.Get("virtual_machine_id").(string)

	// Prepare the VirtualMachineRequest to set the primary IP address
	var vm nb.PatchedVirtualMachineRequest

	if v, ok := d.GetOk("primary_ip4_id"); ok {
		ip4 := v.(string)
		var nullableIP4 nb.NullablePrimaryIPv4
		primaryIPv4 := &nb.PrimaryIPv4{
			Id: &nb.BulkWritableCableRequestStatusId{
				String: stringPtr(ip4),
			},
		}
		nullableIP4.Set(primaryIPv4)
		vm.PrimaryIp4 = nullableIP4
	}

	if v, ok := d.GetOk("primary_ip6_id"); ok {
		ip6 := v.(string)
		var nullableIP6 nb.NullablePrimaryIPv6
		primaryIPv6 := &nb.PrimaryIPv6{
			Id: &nb.BulkWritableCableRequestStatusId{
				String: stringPtr(ip6),
			},
		}
		nullableIP6.Set(primaryIPv6)
		vm.PrimaryIp6 = nullableIP6
	}

	// Update the virtual machine with the primary IP addresses
	_, _, err := c.VirtualizationAPI.VirtualizationVirtualMachinesPartialUpdate(auth, vmID).PatchedVirtualMachineRequest(vm).Execute()
	if err != nil {
		return diag.Errorf("failed to set primary IP address for virtual machine: %s", err.Error())
	}

	// Use VM ID as the resource ID
	d.SetId(vmID)

	return resourcePrimaryIPAddressForVMRead(ctx, d, meta)
}

func resourcePrimaryIPAddressForVMRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*apiClient).Client
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

	vmID := d.Id()

	// Fetch the virtual machine by ID
	vm, _, err := c.VirtualizationAPI.VirtualizationVirtualMachinesRetrieve(auth, vmID).Execute()
	if err != nil {
		d.SetId("")
		return diag.Errorf("failed to read virtual machine: %s", err.Error())
	}

	// Set primary IPs if they exist
	if vm.PrimaryIp4.IsSet() {
		primaryIp4 := vm.PrimaryIp4.Get()
		if primaryIp4 != nil && primaryIp4.Id != nil {
			d.Set("primary_ip4_id", *primaryIp4.Id.String)
		}
	}

	if vm.PrimaryIp6.IsSet() {
		primaryIp6 := vm.PrimaryIp6.Get()
		if primaryIp6 != nil && primaryIp6.Id != nil {
			d.Set("primary_ip6_id", *primaryIp6.Id.String)
		}
	}

	d.Set("virtual_machine_id", vmID)

	return nil
}

func resourcePrimaryIPAddressForVMUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// The update function is identical to the create function since the action is the same
	return resourcePrimaryIPAddressForVMCreate(ctx, d, meta)
}

func resourcePrimaryIPAddressForVMDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*apiClient).Client
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

	vmID := d.Id()

	// Prepare the VirtualMachineRequest to remove the primary IP addresses
	var vm nb.PatchedVirtualMachineRequest
	var nullableIP4 nb.NullablePrimaryIPv4
	var nullableIP6 nb.NullablePrimaryIPv6

	// Set both IPs to null
	nullableIP4.Unset()
	nullableIP6.Unset()

	vm.PrimaryIp4 = nullableIP4
	vm.PrimaryIp6 = nullableIP6

	// Update the virtual machine to unset the primary IPs
	_, _, err := c.VirtualizationAPI.VirtualizationVirtualMachinesPartialUpdate(auth, vmID).PatchedVirtualMachineRequest(vm).Execute()
	if err != nil {
		return diag.Errorf("failed to remove primary IP address for virtual machine: %s", err.Error())
	}

	// Clear the ID from the state
	d.SetId("")

	return nil
}
