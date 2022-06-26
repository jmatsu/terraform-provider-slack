package slack

import "testing"

func Test_Client(t *testing.T) {
	config := &Config{
		Token: "test token",
	}

	client, err := config.ProviderContext()

	if err != nil {
		t.Fatal(err)
	}

	if client.(*Team).client == nil {
		t.Fatalf("required non-nil client")
	}
}
