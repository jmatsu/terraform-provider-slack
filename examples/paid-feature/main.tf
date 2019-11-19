provider "slack" {
    version = "= 0.0.0-snapshot"
    token = var.slack_token
}

data "slack_usergroup" "example" {
    usergroup_id = var.example_data_usergroup_id
}

resource "slack_usergroup" "new" {
    handle = "zz_terraform_example_new"
    name = "New usergroup for terraform example"
}

resource "slack_usergroup" "managed" {
    handle = "zz_terraform_example_managed"
    name = "Managed usergroup for terraform example"
}