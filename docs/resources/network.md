---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "stackit_network Resource - stackit"
subcategory: ""
description: |-
  Manages STACKIT network
  
  -> Environment supportTo set a custom API base URL, set STACKITRESOURCEMANAGEMENT_BASEURL environment variable
---

# stackit_network (Resource)

Manages STACKIT network

<br />

-> __Environment support__<small>To set a custom API base URL, set <code>STACKIT_RESOURCE_MANAGEMENT_BASEURL</code> environment variable </small>



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) the name of the network
- `project_id` (String) The project UUID.

### Optional

- `nameservers` (List of String) List of DNS Servers/Nameservers.
- `prefix_length_v4` (Number) prefix length

### Read-Only

- `network_id` (String) The ID of the network
- `prefixes` (List of String)
- `public_ip` (String) public IP address

