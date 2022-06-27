package slack

import (
	"context"
	"encoding/json"
	"github.com/slack-go/slack"
	"net/http"
	"net/http/httptest"
	"testing"
)

type Routes = []Route

type Route struct {
	Path     string
	Response interface{}
}

func createTestTeam(t *testing.T, routes Routes) (context.Context, *Team) {
	m := http.NewServeMux()

	for _, route := range routes {
		m.HandleFunc(route.Path, func(w http.ResponseWriter, r *http.Request) {
			renderJson(w, route.Response)
		})
	}

	ts := httptest.NewServer(m)

	client := slack.New("test_token",
		slack.Option(slack.OptionHTTPClient(ts.Client())),
		slack.OptionAPIURL(ts.URL+"/"))

	ctx, cancelFunc := context.WithCancel(context.Background())

	t.Cleanup(func() {
		cancelFunc()
		ts.Close()
	})

	return ctx, &Team{
		client: client,
	}
}

func renderJson(w http.ResponseWriter, response interface{}) {
	w.Header().Set("Content-Type", "application/json")
	b, _ := json.Marshal(response)
	w.Write(b)
}
