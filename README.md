# terraform-provider-slack

[![CircleCI](https://circleci.com/gh/jmatsu/terraform-provider-slack.svg?style=svg)](https://circleci.com/gh/jmatsu/terraform-provider-slack)

This is a [Terraform](https://www.terraform.io/) provider for [Slack](https://slack.com)

# Installation

*Recommended way*

Download and put a binary into plugins directory. *e.g. the directory name depends on macOS*

```bash
$ VERSION=<...> curl -sSL "https://raw.githubusercontent.com/jmatsu/terraform-provider-slack/master/scripts/download.sh" | bash
$ mv terraform-provider-slack ~/.terraform.d/plugins/darwin_amd64/
```

Or build a binary by yourself.

```bash
$ go clone ... && cd /path/to/project
$ go mod download
$ go build .
$ mv terraform-provider-slack ~/.terraform.d/plugins/darwin_amd64/
```

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= v0.12.0 (If you are currently using v0.11.x, then please use `v0.1`.)
- Scope: `users:read,usergroups:read,usergroups:write`

## Resources

```hcl
provider "slack" {
  # A token must be of an user. A bot user's token cannot be used for usergroup api call.
  # To get a token, Botkit is one of recommended methods.
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

## Import

```bash
cat<<EOF >> main.tf
resource "slack_usergroup" "foo" {
}

resource "slack_usergroup_members" "bar" {
}
EOF

$ terraform import slack_usergroup.foo <usergroup id>
$ terraform import slack_usergroup_members.bar <usergroup id>
```
