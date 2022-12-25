---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "stackit_kubernetes_project Data Source - stackit"
subcategory: ""
description: |-
  Data source for kubernetes project
---

# stackit_kubernetes_project (Data Source)

Data source for kubernetes project

## Example Usage

```terraform
resource "stackit_kubernetes_project" "example" {
  project_id = "example"
}

data "stackit_kubernetes_project" "example" {
  depends_on = [stackit_kubernetes_project.example]
  project_id = "example"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `project_id` (String) The project ID in which SKE is enabled

### Read-Only

- `id` (String) Specifies the resource ID

