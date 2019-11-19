resource "slack_usergroup" "new" {
    handle = "zz_terraform_example_new"
    name = "New usergroup for terraform example"
}

resource "slack_usergroup" "managed" {
    handle = "zz_terraform_example_managed"
    name = "New usergroup for terraform example"
}