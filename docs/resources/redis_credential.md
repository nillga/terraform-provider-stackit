---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "stackit_redis_credential Resource - stackit"
subcategory: ""
description: |-
  Manages Redis credentials
---

# stackit_redis_credential (Resource)

Manages Redis credentials

## Example Usage

```terraform
resource "stackit_redis_instance" "example" {
  name       = "example"
  project_id = "example"
}
resource "stackit_redis_credential" "example" {
  project_id  = "example"
  instance_id = stackit_redis_instance.example.id
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `instance_id` (String) Redis instance ID the credential belongs to
- `project_id` (String) Project ID the credential belongs to

### Read-Only

- `host` (String) Credential host
- `hosts` (List of String) Credential hosts
- `id` (String) Specifies the resource ID
- `password` (String) Credential password
- `port` (Number) Credential port
- `route_service_url` (String) Credential route service url
- `syslog_drain_url` (String) Credential syslog drain url
- `uri` (String) The instance URI
- `username` (String) Credential username

