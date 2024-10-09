---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "nautobot_vlan Data Source - terraform-provider-nautobot"
subcategory: ""
description: |-
  Retrieves information about a specific VLAN in Nautobot.
---

# nautobot_vlan (Data Source)

Retrieves information about a specific VLAN in Nautobot.



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) The name of the VLAN to retrieve.

### Read-Only

- `created` (String) The creation date of the VLAN.
- `description` (String) Description of the VLAN.
- `id` (String) The UUID of the VLAN.
- `last_updated` (String) The last update date of the VLAN.
- `locations` (List of String) The IDs of the locations associated with the VLAN.
- `role_id` (String) The ID of the role associated with the VLAN.
- `status` (String) The status of the VLAN.
- `tags_ids` (List of String) The IDs of the tags associated with the VLAN.
- `tenant_id` (String) The ID of the tenant associated with the VLAN.
- `vid` (Number) The ID (VID) of the VLAN.
- `vlan_group_id` (String) The ID of the VLAN group.

