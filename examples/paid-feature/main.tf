provider "slack" {
  version = "= 0.0.0-snapshot"
  token   = var.slack_token
}

data "slack_usergroup" "example" {
  usergroup_id = var.example_data_usergroup_id
}

data "slack_user" "sample_1" {
  query_type  = "id"
  query_value = var.example_data_user_id
}

data "slack_user" "sample_2" {
  query_type  = "name"
  query_value = var.example_data_user_name
}

resource "slack_usergroup" "new" {
  handle = "zz_terraform_example_new_${var.salt}"
  name   = "New usergroup for terraform example ${var.salt}"
}

resource "slack_usergroup" "managed" {
  handle = "zz_terraform_example_managed_2"
  name   = "Managed usergroup for terraform example 2"
}

resource "slack_usergroup_members" "example" {
  usergroup_id = slack_usergroup.managed.id
  members      = [data.slack_user.sample_1.id, data.slack_user.sample_2.id]
}

resource "slack_usergroup_channels" "example" {
  usergroup_id = slack_usergroup.managed.id
  channels     = [var.example_data_channel_id]
}