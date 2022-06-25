package slack

import (
	"context"
	"encoding/json"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/slack-go/slack"
	"net/http"
	"net/http/httptest"
	"testing"
)

func respondJson(w http.ResponseWriter, r *http.Request, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	b, _ := json.Marshal(data)
	w.Write(b)
}

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
	m := http.NewServeMux()
	m.HandleFunc("/usergroups.users.list", func(w http.ResponseWriter, r *http.Request) {
		response := userGroupUsersListResponse{
			slack.SlackResponse{Ok: true},
			testUserGroup.Users,
		}
		respondJson(w, r, response)
	})
	ts := httptest.NewServer(m)

	slackClient := slack.New("test_token",
		slack.Option(slack.OptionHTTPClient(ts.Client())),
		slack.OptionAPIURL(ts.URL+"/"))

	team := &Team{slackClient, context.Background()}

	if err := resourceSlackUserGroupMembersRead(d, team); err != nil {
		t.Fatalf("err: %s", err)
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
	m := http.NewServeMux()

	newMembers := &schema.Set{F: schema.HashString}
	for _, u := range testUserGroup.Users {
		newMembers.Add(u)
	}
	if err := d.Set("members", newMembers); err != nil {
		t.Fatalf("err setting existing members: %s", err)
	}
	m.HandleFunc("/usergroups.users.update", func(w http.ResponseWriter, r *http.Request) {
		response := userGroupResponse{
			slack.SlackResponse{Ok: true},
			testUserGroup,
		}
		respondJson(w, r, response)
	})
	ts := httptest.NewServer(m)

	slackClient := slack.New("test_token",
		slack.Option(slack.OptionHTTPClient(ts.Client())),
		slack.OptionAPIURL(ts.URL+"/"))

	team := &Team{slackClient, context.Background()}

	if err := resourceSlackUserGroupMembersCreate(d, team); err != nil {
		t.Fatalf("err: %s", err)
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
	m := http.NewServeMux()
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

	m.HandleFunc("/usergroups.enable", func(w http.ResponseWriter, r *http.Request) {
		response := userGroupResponse{
			slack.SlackResponse{Ok: true},
			testUserGroup,
		}
		respondJson(w, r, response)
	})
	newTestUserGroup := testUserGroup
	newTestUserGroup.Users = append(newTestUserGroup.Users, "NUSERID")

	m.HandleFunc("/usergroups.users.update", func(w http.ResponseWriter, r *http.Request) {
		response := userGroupResponse{
			slack.SlackResponse{Ok: true},
			newTestUserGroup,
		}
		respondJson(w, r, response)
	})
	ts := httptest.NewServer(m)

	slackClient := slack.New("test_token",
		slack.Option(slack.OptionHTTPClient(ts.Client())),
		slack.OptionAPIURL(ts.URL+"/"))

	team := &Team{slackClient, context.Background()}

	if err := resourceSlackUserGroupMembersUpdate(d, team); err != nil {
		t.Fatalf("err update usergroup: %s", err)
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
	m := http.NewServeMux()
	d.SetId(testUserGroup.ID)
	if err := d.Set("usergroup_id", testUserGroup.ID); err != nil {
		t.Fatalf("err set usergroup_id: %s", err)
	}

	m.HandleFunc("/usergroups.disable", func(w http.ResponseWriter, r *http.Request) {
		response := userGroupResponse{
			slack.SlackResponse{Ok: true},
			testUserGroup,
		}
		respondJson(w, r, response)
	})

	ts := httptest.NewServer(m)

	slackClient := slack.New("test_token",
		slack.Option(slack.OptionHTTPClient(ts.Client())),
		slack.OptionAPIURL(ts.URL+"/"))

	team := &Team{slackClient, context.Background()}

	if err := resourceSlackUserGroupMembersDelete(d, team); err != nil {
		t.Fatalf("err update usergroup: %s", err)
	}

	if d.Id() != "" {
		t.Fatalf("expect id to be empty, but got %s", d.Id())
	}

}
