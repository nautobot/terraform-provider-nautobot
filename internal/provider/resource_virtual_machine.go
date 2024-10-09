package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	nb "github.com/nautobot/go-nautobot/v2"
)

func resourceVirtualMachine() *schema.Resource {
	return &schema.Resource{
		Description: "This object manages a virtual machine in Nautobot",

		CreateContext: resourceVirtualMachineCreate,
		ReadContext:   resourceVirtualMachineRead,
		UpdateContext: resourceVirtualMachineUpdate,
		DeleteContext: resourceVirtualMachineDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Virtual Machine's name.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"cluster_id": {
				Description: "Cluster where the virtual machine belongs.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"status": {
				Description: "Status of the virtual machine.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"vcpus": {
				Description: "Number of virtual CPUs.",
				Type:        schema.TypeInt,
				Optional:    true,
			},
			"memory": {
				Description: "Amount of memory in MB.",
				Type:        schema.TypeInt,
				Optional:    true,
			},
			"disk": {
				Description: "Disk size in GB.",
				Type:        schema.TypeInt,
				Optional:    true,
			},
			"comments": {
				Description: "Comments or notes about the virtual machine.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"tenant_id": {
				Description: "Tenant associated with the virtual machine.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"platform_id": {
				Description: "Platform or OS installed on the virtual machine.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"role_id": {
				Description: "Role of the virtual machine.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"primary_ip4_id": {
				Description: "Primary IPv4 address.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"primary_ip6_id": {
				Description: "Primary IPv6 address.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"software_version_id": {
				Description: "Software version installed on the virtual machine.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"software_image_files": {
				Description: "Software image files associated with the software version.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"tags_ids": {
				Description: "Tags associated with the virtual machine.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"created": {
				Description: "Creation date of the virtual machine.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"last_updated": {
				Description: "Last update date of the virtual machine.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func resourceVirtualMachineCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	// Check if the VM with the same name exists
	vmName := d.Get("name").(string)
	existingVMs, _, err := c.VirtualizationAPI.VirtualizationVirtualMachinesList(auth).Name([]string{vmName}).Execute()
	if err != nil {
		return diag.Errorf("failed to list virtual machines: %s", err.Error())
	}

	// If a VM with the same name exists, use its ID and skip creation
	if len(existingVMs.Results) > 0 {
		d.SetId(existingVMs.Results[0].Id)
		return resourceVirtualMachineRead(ctx, d, meta)
	}

	// Convert status name to ID
	statusName := d.Get("status").(string)
	statusID, err := getStatusID(ctx, c, t, statusName)
	if err != nil {
		return diag.Errorf("failed to get status ID for %s: %s", statusName, err.Error())
	}

	// Prepare the VirtualMachineRequest
	var vm nb.VirtualMachineRequest
	vm.Name = d.Get("name").(string)
	vm.Cluster = nb.BulkWritableCableRequestStatus{
		Id: &nb.BulkWritableCableRequestStatusId{
			String: stringPtr(d.Get("cluster_id").(string)),
		},
	}
	vm.Status = nb.BulkWritableCableRequestStatus{
		Id: &nb.BulkWritableCableRequestStatusId{
			String: stringPtr(statusID),
		},
	}

	// Optional fields
	if v, ok := d.GetOk("vcpus"); ok {
		vm.Vcpus.Set(int32Ptr(v.(int)))
	}
	if v, ok := d.GetOk("memory"); ok {
		vm.Memory.Set(int32Ptr(v.(int)))
	}
	if v, ok := d.GetOk("disk"); ok {
		vm.Disk.Set(int32Ptr(v.(int)))
	}
	if v, ok := d.GetOk("comments"); ok {
		comments := v.(string)
		vm.Comments = &comments
	}
	if v, ok := d.GetOk("tenant_id"); ok {
		tenant := v.(string)
		var nullableTenant nb.NullableBulkWritableCircuitRequestTenant
		nullableTenant.Set(&nb.BulkWritableCircuitRequestTenant{
			Id: &nb.BulkWritableCableRequestStatusId{
				String: stringPtr(tenant),
			},
		})
		vm.Tenant = nullableTenant
	}
	if v, ok := d.GetOk("platform_id"); ok {
		platform := v.(string)
		var nullablePlatform nb.NullableBulkWritableCircuitRequestTenant
		nullablePlatform.Set(&nb.BulkWritableCircuitRequestTenant{
			Id: &nb.BulkWritableCableRequestStatusId{
				String: stringPtr(platform),
			},
		})
		vm.Platform = nullablePlatform
	}

	if v, ok := d.GetOk("role_id"); ok {
		role := v.(string)
		var nullableRole nb.NullableBulkWritableCircuitRequestTenant
		nullableRole.Set(&nb.BulkWritableCircuitRequestTenant{
			Id: &nb.BulkWritableCableRequestStatusId{
				String: stringPtr(role),
			},
		})
		vm.Role = nullableRole
	}

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

	if v, ok := d.GetOk("software_version_id"); ok {
		softwareVersion := v.(string)
		var nullableSoftwareVersion nb.NullableBulkWritableVirtualMachineRequestSoftwareVersion
		softwareVersionStruct := &nb.BulkWritableVirtualMachineRequestSoftwareVersion{
			Id: &nb.BulkWritableCableRequestStatusId{
				String: stringPtr(softwareVersion),
			},
		}
		nullableSoftwareVersion.Set(softwareVersionStruct)
		vm.SoftwareVersion = nullableSoftwareVersion
	}

	if v, ok := d.GetOk("software_image_files"); ok {
		var files []nb.SoftwareImageFiles
		for _, file := range v.([]interface{}) {
			fileData := file.(map[string]interface{})
			files = append(files, nb.SoftwareImageFiles{
				Id: &nb.BulkWritableCableRequestStatusId{
					String: stringPtr(fileData["id"].(string)),
				},
			})
		}
		vm.SoftwareImageFiles = files
	}
	if v, ok := d.GetOk("tags_ids"); ok {
		var tags []nb.BulkWritableCableRequestStatus
		for _, tag := range v.([]interface{}) {
			tags = append(tags, nb.BulkWritableCableRequestStatus{
				Id: &nb.BulkWritableCableRequestStatusId{
					String: stringPtr(tag.(string)),
				},
			})
		}
		vm.Tags = tags
	}

	// Create the virtual machine
	rsp, _, err := c.VirtualizationAPI.VirtualizationVirtualMachinesCreate(auth).VirtualMachineRequest(vm).Execute()
	if err != nil {
		return diag.Errorf("failed to create virtual machine: %s", err.Error())
	}

	// Set resource ID
	d.SetId(rsp.Id)

	return resourceVirtualMachineRead(ctx, d, meta)
}

func resourceVirtualMachineRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	// Fetch virtual machine by ID
	vmId := d.Id()
	vm, _, err := c.VirtualizationAPI.VirtualizationVirtualMachinesRetrieve(auth, vmId).Execute()
	if err != nil {
		return diag.Errorf("failed to read virtual machine: %s", err.Error())
	}

	// Map the retrieved data back to Terraform state
	d.Set("name", vm.Name)
	d.Set("cluster_id", vm.Cluster.Id)
	d.Set("status", vm.Status.Id)
	d.Set("vcpus", vm.Vcpus.Get())
	d.Set("memory", vm.Memory.Get())
	d.Set("disk", vm.Disk.Get())
	d.Set("comments", vm.Comments)

	// Handle nullable fields using IsSet and Get methods
	if vm.Tenant.IsSet() {
		tenant := vm.Tenant.Get()
		if tenant != nil && tenant.Id != nil {
			d.Set("tenant_id", *tenant.Id.String)
		}
	}

	if vm.Platform.IsSet() {
		platform := vm.Platform.Get()
		if platform != nil && platform.Id != nil {
			d.Set("platform_id", *platform.Id.String)
		}
	}

	if vm.Role.IsSet() {
		role := vm.Role.Get()
		if role != nil && role.Id != nil {
			d.Set("role_id", *role.Id.String)
		}
	}

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

	if vm.SoftwareVersion.IsSet() {
		softwareVersion := vm.SoftwareVersion.Get()
		if softwareVersion != nil && softwareVersion.Id != nil {
			d.Set("software_version_id", *softwareVersion.Id.String)
		}
	}

	var imageFiles []map[string]string
	for _, file := range vm.SoftwareImageFiles {
		if file.Id != nil && file.Id.String != nil {
			imageFiles = append(imageFiles, map[string]string{
				"id": *file.Id.String,
			})
		}
	}

	d.Set("software_image_files", imageFiles)

	var tags []string
	for _, tag := range vm.Tags {
		if tag.Id != nil {
			tags = append(tags, *tag.Id.String)
		}
	}
	d.Set("tags_ids", tags)

	d.Set("created", vm.Created)
	d.Set("last_updated", vm.LastUpdated)

	return nil
}

func resourceVirtualMachineUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*apiClient).Client
	t := meta.(*apiClient).Token.token

	vmId := d.Id()

	var vm nb.PatchedVirtualMachineRequest

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
		vm.Name = &name
	}

	if d.HasChange("cluster_id") {
		clusterID := d.Get("cluster_id").(string)
		vm.Cluster = &nb.BulkWritableCableRequestStatus{
			Id: &nb.BulkWritableCableRequestStatusId{
				String: &clusterID,
			},
		}
	}

	if d.HasChange("status") {
		statusName := d.Get("status").(string)
		statusID, err := getStatusID(ctx, c, t, statusName)
		if err != nil {
			return diag.Errorf("failed to get status ID for %s: %s", statusName, err.Error())
		}
		vm.Status = &nb.BulkWritableCableRequestStatus{
			Id: &nb.BulkWritableCableRequestStatusId{
				String: stringPtr(statusID),
			},
		}
	}

	// Optional fields
	if d.HasChange("vcpus") {
		vm.Vcpus.Set(int32Ptr(d.Get("vcpus").(int)))
	}
	if d.HasChange("memory") {
		vm.Memory.Set(int32Ptr(d.Get("memory").(int)))
	}
	if d.HasChange("disk") {
		vm.Disk.Set(int32Ptr(d.Get("disk").(int)))
	}
	if d.HasChange("comments") {
		comments := d.Get("comments").(string)
		vm.Comments = &comments
	}
	if d.HasChange("tenant_id") {
		tenant := d.Get("tenant_id").(string)
		var nullableTenant nb.NullableBulkWritableCircuitRequestTenant
		nullableTenant.Set(&nb.BulkWritableCircuitRequestTenant{
			Id: &nb.BulkWritableCableRequestStatusId{
				String: stringPtr(tenant),
			},
		})
		vm.Tenant = nullableTenant
	}
	if d.HasChange("platform_id") {
		platform := d.Get("platform_id").(string)
		var nullablePlatform nb.NullableBulkWritableCircuitRequestTenant
		nullablePlatform.Set(&nb.BulkWritableCircuitRequestTenant{
			Id: &nb.BulkWritableCableRequestStatusId{
				String: stringPtr(platform),
			},
		})
		vm.Platform = nullablePlatform
	}

	if d.HasChange("role_id") {
		role := d.Get("role_id").(string)
		var nullableRole nb.NullableBulkWritableCircuitRequestTenant
		nullableRole.Set(&nb.BulkWritableCircuitRequestTenant{
			Id: &nb.BulkWritableCableRequestStatusId{
				String: stringPtr(role),
			},
		})
		vm.Role = nullableRole
	}

	if d.HasChange("primary_ip4_id") {
		ip4 := d.Get("primary_ip4_id").(string)
		var nullableIP4 nb.NullablePrimaryIPv4
		primaryIPv4 := &nb.PrimaryIPv4{
			Id: &nb.BulkWritableCableRequestStatusId{
				String: stringPtr(ip4),
			},
		}
		nullableIP4.Set(primaryIPv4)
		vm.PrimaryIp4 = nullableIP4
	}

	if d.HasChange("primary_ip6_id") {
		ip6 := d.Get("primary_ip6_id").(string)
		var nullableIP6 nb.NullablePrimaryIPv6
		primaryIPv6 := &nb.PrimaryIPv6{
			Id: &nb.BulkWritableCableRequestStatusId{
				String: stringPtr(ip6),
			},
		}
		nullableIP6.Set(primaryIPv6)
		vm.PrimaryIp6 = nullableIP6
	}

	if d.HasChange("software_version_id") {
		softwareVersion := d.Get("software_version_id").(string)
		var nullableSoftwareVersion nb.NullableBulkWritableVirtualMachineRequestSoftwareVersion
		softwareVersionStruct := &nb.BulkWritableVirtualMachineRequestSoftwareVersion{
			Id: &nb.BulkWritableCableRequestStatusId{
				String: stringPtr(softwareVersion),
			},
		}
		nullableSoftwareVersion.Set(softwareVersionStruct)
		vm.SoftwareVersion = nullableSoftwareVersion
	}

	if d.HasChange("software_image_files") {
		var files []nb.SoftwareImageFiles
		for _, file := range d.Get("software_image_files").([]interface{}) {
			fileData := file.(map[string]interface{})
			files = append(files, nb.SoftwareImageFiles{
				Id: &nb.BulkWritableCableRequestStatusId{
					String: stringPtr(fileData["id"].(string)),
				},
			})
		}
		vm.SoftwareImageFiles = files
	}

	if d.HasChange("tags_ids") {
		var tags []nb.BulkWritableCableRequestStatus
		for _, tag := range d.Get("tags_ids").([]interface{}) {
			tagID := tag.(string)
			tags = append(tags, nb.BulkWritableCableRequestStatus{
				Id: &nb.BulkWritableCableRequestStatusId{
					String: &tagID,
				},
			})
		}
		vm.Tags = tags
	}

	// Call the API to update the virtual machine
	_, _, err := c.VirtualizationAPI.VirtualizationVirtualMachinesPartialUpdate(auth, vmId).PatchedVirtualMachineRequest(vm).Execute()
	if err != nil {
		return diag.Errorf("failed to update virtual machine: %s", err.Error())
	}

	return resourceVirtualMachineRead(ctx, d, meta)
}

func resourceVirtualMachineDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	// Delete the virtual machine by ID
	vmId := d.Id()
	_, err := c.VirtualizationAPI.VirtualizationVirtualMachinesDestroy(auth, vmId).Execute()
	if err != nil {
		return diag.Errorf("failed to delete virtual machine: %s", err.Error())
	}

	// Clear the ID
	d.SetId("")

	return nil
}
