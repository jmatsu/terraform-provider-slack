# terraform-provider-slack

[![CircleCI](https://circleci.com/gh/jmatsu/terraform-provider-slack.svg?style=svg)](https://circleci.com/gh/jmatsu/terraform-provider-slack)

This is a [Terraform](https://www.terraform.io/) provider for [Slack](https://slack.com)

# Installation

*Recommended way*

Download and put a binary into plugins directory. *e.g. the directory name depends on macOS*

```bash
$ export VERSION=<...>
$ curl -sSL "https://raw.githubusercontent.com/jmatsu/terraform-provider-slack/master/scripts/download.sh" | bash
$ mv terraform-provider-slack_$VERSION ~/.terraform.d/plugins/darwin_amd64/
```

Or build a binary by yourself.

```bash
$ go clone ... && cd /path/to/project
$ go mod download
$ go build .
$ mv terraform-provider-slack ~/.terraform.d/plugins/darwin_amd64/
```

See https://www.terraform.io/docs/configuration/providers.html#third-party-plugins for more details.

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= v0.12.0 (v0.11.x may work but not supported actively)
- Scope: `users:read,usergroups:read,usergroups:write,channels:read,channels:write,groups:read,groups:write` ref [bot.d/src/bot.ts](./bot.d/src/bot.ts)

## Limitations

**I do not have any Plus or Enterprise Grid workspace which I'm free to use unfotunately.**

That's why several resources, e.g. a slack user, have not been supported yet. 

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

data "slack_channel" "..." {
  channel_id = <channel id>
}

data "slack_group" "..." {
  group_id = <group id>
}

data "slack_usergroup" "..." {
  usergroup_id = <usergroup id>
}

resource "slack_channel" "..." {
  name = "<name>"
  topic = "..."
  purpose = "..."
  is_archive = <true|false>
}

resource "slack_group" "..." {
  name = "<name>"
  topic = "..."
  purpose = "..."
  is_archive = <true|false>
}

resource "slack_usergroup" "..." {
  handle      = "<mention name>"
  name        = "<name>"
  description = "..."
  auto_type   = "" or "admins" or "owners"
}

resource "slack_usergroup_members" "..." {
  usergroup_id = "<usergroup id>"
  members = ["<user id>", ...]
}

resource "slack_usergroup_channels" "..." {
  usergroup_id = "<usergroup id>"
  channels = ["<channel id>", ...]
}
```

## Import

```bash
$ terraform import slack_channel.<name> <channel id>
$ terraform import slack_group.<name> <group id>
$ terraform import slack_usergroup.<name> <usergroup id>
$ terraform import slack_usergroup_members.<name> <usergroup id>
$ terraform import slack_usergroup_channels.<name> <usergroup id>
```
