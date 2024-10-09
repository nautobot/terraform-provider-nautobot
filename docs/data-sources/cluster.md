---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "nautobot_cluster Data Source - terraform-provider-nautobot"
subcategory: ""
description: |-
  Retrieves information about a specific cluster in Nautobot.
---

# nautobot_cluster (Data Source)

Retrieves information about a specific cluster in Nautobot.



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) The name of the cluster.

### Read-Only

- `cluster_group_id` (String) The ID of the cluster group.
- `cluster_type_id` (String) The ID of the cluster type.
- `comments` (String) Comments or notes about the cluster.
- `created` (String) The creation date of the cluster.
- `id` (String) The UUID of the cluster.
- `last_updated` (String) The last update date of the cluster.
- `location_id` (String) The ID of the location associated with the cluster.
- `tags_ids` (List of String) The IDs of the tags associated with the cluster.
- `tenant_id` (String) The ID of the tenant associated with the cluster.

