# terraform-provider-slack

[![CircleCI](https://circleci.com/gh/jmatsu/terraform-provider-slack.svg?style=svg)](https://circleci.com/gh/jmatsu/terraform-provider-slack)

This is a [Terraform](https://www.terraform.io/) provider for [Slack](https://slack.com)

# Installation

ref: https://registry.terraform.io/providers/jmatsu/slack/latest

Or build a binary by yourself.

```bash
$ go clone ... && cd /path/to/project
$ go mod download
$ go build .
$ mv terraform-provider-slack ~/.terraform.d/plugins/[architecture name]/
```

See https://www.terraform.io/docs/configuration/providers.html#third-party-plugins for more details.

# Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= v0.12.0 (v0.11.x may work but not supported actively)
- Scope: `users:read,users:read.email,usergroups:read,usergroups:write,channels:read,channels:write,groups:read,groups:write`
  - `users:read.email` is required since v0.6.0

# Limitations

Several resources that require Slack Plus or Enterprise Grid are not supported. e.g. a slack user (not a data source)

# Resources

```hcl
provider "slack" {
  # A token must be of an user. A bot user's token cannot be used for usergroup api call.
  # To get a token, Botkit is one of recommended methods.
  token = "SLACK_TOKEN"
}

data "slack_user" "..." {
  query_type = "id" or "email"
  query_value = "<user id>" or "<email>"
}

data "slack_conversation" "..." {
  channel_id = <channel id>
}

data "slack_usergroup" "..." {
  usergroup_id = <usergroup id>
}

resource "slack_conversation" "..." {
  name = "<name>"
  topic = "..."
  purpose = "..."
  action_on_destroy = "<archive|none>" # this is required since v0.8.0
  is_archive = <true|false>
  is_private = <true|false>
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

# Import

```bash
$ terraform import slack_conversation.<name> <channel id>
$ terraform import slack_usergroup.<name> <usergroup id>
$ terraform import slack_usergroup_members.<name> <usergroup id>
$ terraform import slack_usergroup_channels.<name> <usergroup id>
```

# Trouble Shooting

## Show provider's debug logs

Enable the provider logging. All custom logs should have `provider=jmatsu/slack` in its body.

```bash
# The following expects you are using `slack` as provider name
TF_LOG=json TF_LOG_PROVIDER_SLACK=TRACE ...cmd
```

# Development

Source codes consist of an entrypoint `main.go` and `slack/**.go`.

## Try the built provider in projects

Build and install the provider into your machine.

> Currently, only `~/.terraform.d` is supported.

```bash
# choose OS_NAME and ARCH of your machine
make install-${OS_NAME}_${ARCH}
```

Use the built provider by specifying the custom source.

```terraform
terraform {
  required_providers {
    slack = {
      # please make sure ~/.terraform.d/plugins/github.com/jmatsu/slack/${version}/${OS_NAME}_${ARCH}/terraform-provider-slack_v${version} exists and it's executable.
      source = "github.com/jmatsu/slack"
      version = "${version}"
    }
  }
}
```

## Other commands

```
# Build the binary
make build

# Run test
make test

# Apply gofmt (overwrite mode)
make fmt
```

# Release

CI will build and archive the release artifacts to GitHub Releases and terraform provider registry. 

```
version="/\d\.\d\.\d/"

# please make sure your working branch is same to the default branch.
git tag "v$version"
git push "v$version"
```

# LICENSE

Under [MIT](./LICENSE)

# Maintainers

@jmatsu, @billcchung