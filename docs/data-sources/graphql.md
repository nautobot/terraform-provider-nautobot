---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "nautobot_graphql Data Source - terraform-provider-nautobot"
subcategory: ""
description: |-
  Provide an interface to make GraphQL calls to Nautobot as a flexible data source.
---

# nautobot_graphql (Data Source)

Provide an interface to make GraphQL calls to Nautobot as a flexible data source.



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `query` (String) The GraphQL query that will be sent to Nautobot.

### Read-Only

- `data` (String) The data returned by the GraphQL query.
- `id` (String) The ID of this resource.


