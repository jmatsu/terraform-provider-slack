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

data "slack_user" "sample_3" {
  query_type  = "name"
  query_value = var.example_data_user_email
}

resource "slack_usergroup" "new" {
  handle = "zz_terraform_example_new_${var.salt}"
  name   = "New usergroup for terraform example ${var.salt}"
}

resource "slack_usergroup" "managed" {
  handle = "zz_terraform_example_managed_2"
  name   = "Managed usergroup for terraform example 2"
}

resource "slack_usergroup" "new2" {
  handle = "zz_terraform_example_new_x"
  name   = "New usergroup for terraform example x"
}

resource "slack_usergroup_channels" "example" {
  usergroup_id = slack_usergroup.managed.id
  channels     = [var.example_data_channel_id]
}