package slack

import (
	"encoding/json"
	"github.com/slack-go/slack"
	"net/http"
	"net/http/httptest"
)

type Routes = []Route

type Route struct {
	Path     string
	Response interface{}
}

func createStubClient(routes Routes) *slack.Client {
	m := http.NewServeMux()

	for _, route := range routes {
		m.HandleFunc(route.Path, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			b, _ := json.Marshal(route.Response)
			w.Write(b)
		})
	}

	ts := httptest.NewServer(m)

	return slack.New("test_token",
		slack.Option(slack.OptionHTTPClient(ts.Client())),
		slack.OptionAPIURL(ts.URL+"/"))
}
