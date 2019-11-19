provider "slack" {
  version = "= 0.0.0-snapshot"
  token   = var.slack_token
}

data "slack_user" "sample_1" {
  query_type  = "id"
  query_value = var.example_data_user_id
}

data "slack_user" "sample_2" {
  query_type  = "name"
  query_value = var.example_data_user_name
}

data "slack_channel" "existing_sample_1" {
  channel_id = var.example_data_channel_id
}

data "slack_group" "existing_sample_2" {
  group_id = var.example_data_group_id
}

resource "slack_channel" "new" {
  name = "zz_terraform_example_channel_new-${var.salt}}"
}

resource "slack_group" "new" {
  name = "zz_terraform_example_group_new-${var.salt}"
}

resource "slack_channel" "managed" {
  name    = "zz_terraform_example_channel_managed-1"
  purpose = "change here"
}

resource "slack_group" "managed" {
  name    = "zz_terraform_example_group_managed-1"
  purpose = "change here"
}
