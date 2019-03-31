# Terraform Provider for Slack

This is a [Terraform](https://www.terraform.io/) provider for [Slack](https://slack.com)

# Installation

```bash
$ VERSION=<...> curl -sSL "https://raw.githubusercontent.com/jmatsu/terraform-provider-slack/master/scripts/download.sh" | bash
```

```bash
$ go get https://github.com/jmatsu/terraform-provider-slack
```

```bash
$ go clone ... && cd /path/to/project
$ go mod download
$ go install
```

## Requirements

`[Terraform](https://www.terraform.io/downloads.html) >= v0.11.0`

## Resources

```hcl
provider "slack" {
  token = "SLACK_TOKEN"
}

data "slack_user" "..." {
  query_type = "name" or "id"
  query_value = "<name or real name>" or "<user id>"
}

resource "slack_usergroup" "..." {
  handle      = "<mention name>"
  name        = "<name>"
  description = "..."
  auto_type   = "" or "admins" or "owners"
}

resource "slack_usergroup_members" "..." {
  usergroup_id = "<usergroup id>"
  members = ["<user id>"]
}
```

