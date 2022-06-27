package slack

import "testing"

func Test_Client(t *testing.T) {
	config := &Config{
		Token: "test token",
	}

	team, err := config.ProviderContext("version", "commit")

	if err != nil {
		t.Fatal(err)
	}

	if team.client == nil {
		t.Fatalf("required non-nil client")
	}
}
