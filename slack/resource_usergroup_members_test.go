package slack

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/slack-go/slack"
	"testing"
)

func stringInSlice(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

type userGroupResponse struct {
	slack.SlackResponse
	UserGroup slack.UserGroup `json:"usergroup"`
}

type userGroupUsersListResponse struct {
	slack.SlackResponse
	Users []string `json:"users"`
}

var testUserGroup = slack.UserGroup{
	ID:    "S0615G0KT",
	Users: []string{"U0614TZR7", "U060RNRCZ"},
}

func Test_ResourceUserGroupMembersRead(t *testing.T) {
	d := resourceSlackUserGroupMembers().TestResourceData()
	ctx, team := createTestTeam(t, Routes{
		{
			Path: "/usergroups.users.list",
			Response: userGroupUsersListResponse{
				slack.SlackResponse{Ok: true},
				testUserGroup.Users,
			},
		},
	})

	if diags := resourceSlackUserGroupMembersRead(ctx, d, team); diags.HasError() {
		for _, d := range diags {
			if d.Severity == diag.Error {
				t.Fatalf("err: %s", d.Summary)
			}
		}
	}

	members := d.Get("members").(*schema.Set)
	if len(testUserGroup.Users) != members.Len() {
		t.Fatalf("expect %v members but got %v", len(testUserGroup.Users), members.Len())
	}
	for _, m := range members.List() {
		if !stringInSlice(testUserGroup.Users, m.(string)) {
			t.Fatalf("unexpected user ID %s", m)
		}
	}
}

func Test_ResourceUserGroupMembersCreate(t *testing.T) {
	d := resourceSlackUserGroupMembers().TestResourceData()

	newMembers := &schema.Set{F: schema.HashString}
	for _, u := range testUserGroup.Users {
		newMembers.Add(u)
	}
	if err := d.Set("members", newMembers); err != nil {
		t.Fatalf("err setting existing members: %s", err)
	}

	ctx, team := createTestTeam(t, Routes{
		{
			Path: "/usergroups.users.update",
			Response: userGroupResponse{
				slack.SlackResponse{Ok: true},
				testUserGroup,
			},
		},
	})

	if diags := resourceSlackUserGroupMembersCreate(ctx, d, team); diags.HasError() {
		for _, d := range diags {
			if d.Severity == diag.Error {
				t.Fatalf("err: %s", d.Summary)
			}
		}
	}

	members := d.Get("members").(*schema.Set)
	if len(testUserGroup.Users) != members.Len() {
		t.Fatalf("expect %v members but got %v", len(testUserGroup.Users), members.Len())
	}
	for _, m := range members.List() {
		if !stringInSlice(testUserGroup.Users, m.(string)) {
			t.Fatalf("unexpected user ID %s", m)
		}
	}
}

func Test_ResourceUserGroupMembersUpdate(t *testing.T) {
	d := resourceSlackUserGroupMembers().TestResourceData()
	d.SetId(testUserGroup.ID)
	if err := d.Set("usergroup_id", testUserGroup.ID); err != nil {
		t.Fatalf("err set usergroup_id: %s", err)
	}
	existingMembers := &schema.Set{F: schema.HashString}
	for _, u := range testUserGroup.Users {
		existingMembers.Add(u)
	}
	if err := d.Set("members", existingMembers); err != nil {
		t.Fatalf("err setting existing members: %s", err)
	}

	newTestUserGroup := testUserGroup
	newTestUserGroup.Users = append(newTestUserGroup.Users, "NUSERID")

	ctx, team := createTestTeam(t, Routes{
		{
			Path:     "/usergroups.enable",
			Response: slack.SlackResponse{Ok: true},
		},
		{
			Path: "/usergroups.users.update",
			Response: userGroupResponse{
				slack.SlackResponse{Ok: true},
				newTestUserGroup,
			},
		},
	})

	if diags := resourceSlackUserGroupMembersUpdate(ctx, d, team); diags.HasError() {
		for _, d := range diags {
			if d.Severity == diag.Error {
				t.Fatalf("err: %s", d.Summary)
			}
		}
	}

	members := d.Get("members").(*schema.Set)
	if len(newTestUserGroup.Users) != members.Len() {
		t.Fatalf("expect %v members but got %v", len(newTestUserGroup.Users), members.Len())
	}
	for _, m := range members.List() {
		if !stringInSlice(newTestUserGroup.Users, m.(string)) {
			t.Fatalf("unexpected user ID %s", m)
		}
	}
}

func Test_ResourceUserGroupMembersDelete(t *testing.T) {
	d := resourceSlackUserGroupMembers().TestResourceData()
	d.SetId(testUserGroup.ID)
	if err := d.Set("usergroup_id", testUserGroup.ID); err != nil {
		t.Fatalf("err set usergroup_id: %s", err)
	}

	ctx, team := createTestTeam(t, Routes{
		{
			Path: "/usergroups.disable",
			Response: userGroupResponse{
				slack.SlackResponse{Ok: true},
				testUserGroup,
			},
		},
	})

	if diags := resourceSlackUserGroupMembersDelete(ctx, d, team); diags.HasError() {
		for _, d := range diags {
			if d.Severity == diag.Error {
				t.Fatalf("err: %s", d.Summary)
			}
		}
	}

	if d.Id() != "" {
		t.Fatalf("expect id to be empty, but got %s", d.Id())
	}

}
