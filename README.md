# Terraform Provider for STACKIT
<!--summary-image-->
<img src="https://hcti.io/v1/image/250055f2-e4b8-4320-85c2-1068002e63c0" width="250" align="right" />
<!--revision-af875b11-1ff4-4c36-a307-651d4ce510ca--><!--summary-image-->

[![Go Report Card](https://goreportcard.com/badge/github.com/SchwarzIT/terraform-provider-stackit)](https://goreportcard.com/report/github.com/SchwarzIT/terraform-provider-stackit) <!--workflow-badge-->[![GitHub Workflow Status](https://img.shields.io/badge/Acceptance%20Tests-All%20failed-red)](https://github.com/SchwarzIT/terraform-provider-stackit/actions/workflows/acceptance_test.yml)<!--revision-d2321bec-c8cd-47ac-8918-310ee80cb83d--><!--workflow-badge--><br />[![GitHub release (latest by date)](https://img.shields.io/github/v/release/SchwarzIT/terraform-provider-stackit)](https://registry.terraform.io/providers/SchwarzIT/stackit/latest/docs) [![GitHub go.mod Go version of a Go module](https://img.shields.io/github/go-mod/go-version/gomods/athens.svg)](https://github.com/gomods/athens) [![License](https://img.shields.io/badge/License-Apache_2.0-lightgray.svg)](https://opensource.org/licenses/Apache-2.0)

The STACKIT provider is a project developed and maintained by the STACKIT community within Schwarz IT.

**⚠️ Important:** This project will continue to be maintained as long as internal teams are using it.

> Note that the official STACKIT provider can be found [here](https://github.com/stackitcloud/terraform-provider-stackit).<br />To migrate to the official provider from the community provider, [refer to the migration notes](https://github.com/stackitcloud/terraform-provider-stackit/blob/main/MIGRATION.md).

📖 [Provider Documentation](https://registry.terraform.io/providers/SchwarzIT/stackit/latest/docs)

🚀 [STACKIT Community Go Client](https://github.com/SchwarzIT/community-stackit-go-client)

&nbsp;

## Reporting Issues

If you encounter any issues or have suggestions for improvements, please open an issue in the repository. We appreciate your feedback and will do our best to address any problems as soon as possible.

The provider is built and maintained by the STACKIT community in Schwarz IT and is not an official STACKIT provider

&nbsp;

## Usage Example

```hcl

terraform {
  required_providers {
    stackit = {
      source  = "SchwarzIT/stackit"
      version = "~> 1.27.0"
    }
  }
}

# Configure the STACKIT Provider
provider "stackit" {
  service_account_key_path = var.service_account_key_path
  private_key_path         = var.private_key_path
}

```

For further authentication methods, please refer to our [Provider Documentation](https://registry.terraform.io/providers/SchwarzIT/stackit/latest/docs)
