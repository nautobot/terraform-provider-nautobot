package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	nb "github.com/nautobot/go-nautobot/v2"
)

func resourceVMInterface() *schema.Resource {
	return &schema.Resource{
		Description: "This object manages a VM Interface in Nautobot",

		CreateContext: resourceVMInterfaceCreate,
		ReadContext:   resourceVMInterfaceRead,
		UpdateContext: resourceVMInterfaceUpdate,
		DeleteContext: resourceVMInterfaceDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Name of the VM interface.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"mac_address": {
				Description: "MAC address of the interface.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"enabled": {
				Description: "Whether the interface is enabled.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
			},
			"mtu": {
				Description: "MTU size of the interface.",
				Type:        schema.TypeInt,
				Optional:    true,
			},
			"mode": {
				Description: "Mode of the interface.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"description": {
				Description: "Description of the interface.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"status": {
				Description: "Status of the VM interface.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"virtual_machine_id": {
				Description: "ID of the virtual machine to which the interface belongs.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"untagged_vlan_id": {
				Description: "Untagged VLAN ID associated with the interface.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"tags_ids": {
				Description: "Tags associated with the interface.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"ip_addresses": {
				Description: "List of IP addresses to assign to the VM interface.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"created": {
				Description: "Creation date of the interface.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"last_updated": {
				Description: "Last updated date of the interface.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func resourceVMInterfaceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	// Check if the interface with the same name and virtual machine ID exists
	interfaceName := d.Get("name").(string)
	virtualMachineID := d.Get("virtual_machine_id").(string)

	existingInterfaces, _, err := c.VirtualizationAPI.VirtualizationInterfacesList(auth).
		Name([]string{interfaceName}).
		VirtualMachine([]string{virtualMachineID}).
		Execute()
	if err != nil {
		return diag.Errorf("failed to list VM interfaces: %s", err.Error())
	}

	if len(existingInterfaces.Results) > 0 {
		// Interface already exists, use its ID and skip creation
		d.SetId(existingInterfaces.Results[0].Id)
		return resourceVMInterfaceRead(ctx, d, meta)
	}

	// Convert status name to ID
	statusName := d.Get("status").(string)
	statusID, err := getStatusID(ctx, c, t, statusName)
	if err != nil {
		return diag.Errorf("failed to get status ID for %s: %s", statusName, err.Error())
	}

	// Prepare the VMInterface request
	var vmInterface nb.WritableVMInterfaceRequest
	vmInterface.Name = interfaceName
	vmInterface.Status = nb.BulkWritableCableRequestStatus{
		Id: &nb.BulkWritableCableRequestStatusId{
			String: &statusID,
		},
	}

	// Set optional fields
	if v, ok := d.GetOk("mac_address"); ok {
		vmInterface.MacAddress.Set(stringPtr(v.(string)))
	}
	if v, ok := d.GetOk("enabled"); ok {
		enabled := v.(bool)
		vmInterface.Enabled = &enabled
	}
	if v, ok := d.GetOk("mtu"); ok {
		mtu := int32(v.(int))
		vmInterface.Mtu.Set(&mtu)
	}
	if v, ok := d.GetOk("description"); ok {
		desc := v.(string)
		vmInterface.Description = &desc
	}

	// Handle virtual machine ID
	vmInterface.VirtualMachine.Id = &nb.BulkWritableCableRequestStatusId{
		String: &virtualMachineID,
	}

	// Create the interface
	rsp, _, err := c.VirtualizationAPI.VirtualizationInterfacesCreate(auth).WritableVMInterfaceRequest(vmInterface).Execute()
	if err != nil {
		return diag.Errorf("failed to create VM interface: %s", err.Error())
	}

	// Set resource ID
	d.SetId(rsp.Id)

	// Assign IP addresses to the VM interface
	if v, ok := d.GetOk("ip_addresses"); ok {
		ipAddresses := v.([]interface{})
		for _, ip := range ipAddresses {
			err := assignIPAddressToVMInterface(ctx, c, t, ip.(string), rsp.Id)
			if err != nil {
				return diag.Errorf("failed to assign IP address to VM interface: %s", err.Error())
			}
		}
	}

	return resourceVMInterfaceRead(ctx, d, meta)
}

func resourceVMInterfaceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	// Fetch interface by ID
	vmInterfaceId := d.Id()
	vmInterface, _, err := c.VirtualizationAPI.VirtualizationInterfacesRetrieve(auth, vmInterfaceId).Execute()
	if err != nil {
		return diag.Errorf("failed to read VM interface: %s", err.Error())
	}

	// Map the retrieved data back to Terraform state
	d.Set("name", vmInterface.Name)
	d.Set("mac_address", vmInterface.MacAddress)
	d.Set("enabled", vmInterface.Enabled)
	d.Set("mtu", vmInterface.Mtu)
	d.Set("description", vmInterface.Description)
	d.Set("status", vmInterface.Status.Id)
	d.Set("virtual_machine_id", vmInterface.VirtualMachine.Id)
	d.Set("untagged_vlan_id", vmInterface.UntaggedVlan)
	d.Set("created", vmInterface.Created)
	d.Set("last_updated", vmInterface.LastUpdated)
	d.Set("tags_ids", vmInterface.Tags)

	// Fetch assigned IP addresses
	assignedIPs := []string{}
	for _, ip := range vmInterface.IpAddresses {
		assignedIPs = append(assignedIPs, *ip.Id.String)
	}
	d.Set("ip_addresses", assignedIPs)

	if vmInterface.Mode != nil {
		d.Set("mode", vmInterface.Mode.Label)
	}

	return nil
}

func resourceVMInterfaceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*apiClient).Client
	t := meta.(*apiClient).Token.token

	vmInterfaceId := d.Id()

	var vmInterface nb.PatchedWritableVMInterfaceRequest

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

	// Update the fields that have changed
	if d.HasChange("name") {
		name := d.Get("name").(string)
		vmInterface.Name = &name
	}
	if d.HasChange("mac_address") {
		mac := d.Get("mac_address").(string)
		vmInterface.MacAddress.Set(&mac)
	}
	if d.HasChange("enabled") {
		enabled := d.Get("enabled").(bool)
		vmInterface.Enabled = &enabled
	}
	if d.HasChange("mtu") {
		mtu := int32(d.Get("mtu").(int))
		vmInterface.Mtu.Set(&mtu)
	}
	if d.HasChange("description") {
		description := d.Get("description").(string)
		vmInterface.Description = &description
	}
	if d.HasChange("status") {
		statusName := d.Get("status").(string)
		statusID, err := getStatusID(ctx, c, t, statusName)
		if err != nil {
			return diag.Errorf("failed to get status ID for %s: %s", statusName, err.Error())
		}
		vmInterface.Status = &nb.BulkWritableCableRequestStatus{
			Id: &nb.BulkWritableCableRequestStatusId{
				String: &statusID,
			},
		}
	}
	if d.HasChange("virtual_machine_id") {
		vmID := d.Get("virtual_machine_id").(string)
		vmInterface.VirtualMachine = &nb.BulkWritableCableRequestStatus{
			Id: &nb.BulkWritableCableRequestStatusId{
				String: &vmID,
			},
		}
	}

	// Call the API to update the VM interface
	_, _, err := c.VirtualizationAPI.VirtualizationInterfacesPartialUpdate(auth, vmInterfaceId).PatchedWritableVMInterfaceRequest(vmInterface).Execute()
	if err != nil {
		return diag.Errorf("failed to update VM interface: %s", err.Error())
	}

	// Update IP addresses if they have changed
	if d.HasChange("ip_addresses") {
		oldIPs, newIPs := d.GetChange("ip_addresses")

		// Remove old IP addresses
		for _, oldIP := range oldIPs.([]interface{}) {
			err := removeIPAddressFromVMInterface(ctx, c, t, oldIP.(string), vmInterfaceId) // Pass vmInterfaceId here
			if err != nil {
				return diag.Errorf("failed to remove IP address from VM interface: %s", err.Error())
			}
		}

		// Assign new IP addresses
		for _, newIP := range newIPs.([]interface{}) {
			err := assignIPAddressToVMInterface(ctx, c, t, newIP.(string), vmInterfaceId)
			if err != nil {
				return diag.Errorf("failed to assign IP address to VM interface: %s", err.Error())
			}
		}
	}

	return resourceVMInterfaceRead(ctx, d, meta)
}

func resourceVMInterfaceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	// Delete the interface by ID
	vmInterfaceId := d.Id()
	_, err := c.VirtualizationAPI.VirtualizationInterfacesDestroy(auth, vmInterfaceId).Execute()
	if err != nil {
		return diag.Errorf("failed to delete VM interface: %s", err.Error())
	}

	// Clear the ID
	d.SetId("")

	return nil
}

// Helper function to assign an IP address to a VM interface
func assignIPAddressToVMInterface(ctx context.Context, c *nb.APIClient, token, ipAddressID, vmInterfaceID string) error {
	// Auth context
	auth := context.WithValue(
		ctx,
		nb.ContextAPIKeys,
		map[string]nb.APIKey{
			"tokenAuth": {
				Key:    token,
				Prefix: "Token",
			},
		},
	)

	// Wrap the ipAddressID and vmInterfaceID in BulkWritableCableRequestStatusId
	ipAddressStatusId := nb.BulkWritableCableRequestStatusId{String: &ipAddressID}
	vmInterfaceStatusId := nb.BulkWritableCableRequestStatusId{String: &vmInterfaceID}

	// Create BulkWritableCircuitRequestTenant for the VM Interface
	vmInterfaceTenant := nb.BulkWritableCircuitRequestTenant{
		Id: &vmInterfaceStatusId,
	}

	// Create the NullableBulkWritableCircuitRequestTenant as a value (not a pointer)
	vmInterfaceNullableTenant := nb.NullableBulkWritableCircuitRequestTenant{}
	vmInterfaceNullableTenant.Set(&vmInterfaceTenant)

	// Prepare the request to assign the IP address to the VM interface
	ipToInterfaceRequest := nb.IPAddressToInterfaceRequest{
		IpAddress: nb.BulkWritableCableRequestStatus{
			Id: &ipAddressStatusId, // Assign the IP address ID
		},
		// Properly set VmInterface using NullableBulkWritableCircuitRequestTenant
		VmInterface: vmInterfaceNullableTenant, // Use the value, not a pointer
	}

	// Call the API to link the IP address with the VM interface
	_, _, err := c.IpamAPI.IpamIpAddressToInterfaceCreate(auth).IPAddressToInterfaceRequest(ipToInterfaceRequest).Execute()
	if err != nil {
		return fmt.Errorf("failed to assign IP address to VM interface: %v", err)
	}

	return nil
}

// Helper function to remove an IP address from a VM interface
func removeIPAddressFromVMInterface(ctx context.Context, c *nb.APIClient, token, ipAddressID, vmInterfaceID string) error {
	// Auth context
	auth := context.WithValue(
		ctx,
		nb.ContextAPIKeys,
		map[string]nb.APIKey{
			"tokenAuth": {
				Key:    token,
				Prefix: "Token",
			},
		},
	)

	// Retrieve the IP address object to find the related VM interface assignment
	ipAddress, _, err := c.IpamAPI.IpamIpAddressesRetrieve(auth, ipAddressID).Execute()
	if err != nil {
		return fmt.Errorf("failed to retrieve IP address: %v", err)
	}

	// Look for the specific VM interface assignment in the IP address object
	var assignmentID string
	for _, vmInterface := range ipAddress.VmInterfaces {
		if vmInterface.Id.String != nil && *vmInterface.Id.String == vmInterfaceID {
			assignmentID = *vmInterface.Id.String
			break
		}
	}

	// If no assignment is found, return an error
	if assignmentID == "" {
		return fmt.Errorf("no assignment found for IP address %s and VM interface %s", ipAddressID, vmInterfaceID)
	}

	// Call IpamIpAddressToInterfaceDestroy to remove the assignment
	_, err = c.IpamAPI.IpamIpAddressToInterfaceDestroy(auth, assignmentID).Execute()
	if err != nil {
		return fmt.Errorf("failed to remove IP address assignment: %v", err)
	}

	return nil
}
