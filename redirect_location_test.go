// Package traefik_plugin_redirect_ipv6 is a traefik plugin fixing the location header in a redirect response
package redirect_ipv6 //nolint
import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)


func TestDefaultHandling(t *testing.T) {
	tests := []struct {
		desc            string
		forwardedPrefix string
		forwardedHost   string
		locationBefore  string
		expLocation     string
	}{
		{
			desc:           "No forwarded Prefix and relative path",
			locationBefore: "somevalue",
			expLocation:    "somevalue",
		},
		{
			desc:           "No forwarded Prefix and absolute path",
			locationBefore: "http://host:815/path",
			expLocation:    "http://host:815/path",
		},
	}

	config := &Config{
		Default:  true,
		Rewrites: make([]Rewrite, 0),
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			next := func(rw http.ResponseWriter, req *http.Request) {
				rw.Header().Add("Location", test.locationBefore)
				rw.WriteHeader(http.StatusMovedPermanently)
			}

			redirectLocation, err := New(context.Background(), http.HandlerFunc(next), config, "redirectLocation")
			if err != nil {
				t.Fatal(err)
			}

			recorder := httptest.NewRecorder()

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if len(test.forwardedPrefix) > 0 {
				req.Header.Add("X-Forwarded-Prefix", test.forwardedPrefix)
			}
			if len(test.forwardedHost) > 0 {
				req.Header.Add("X-Forwarded-Host", test.forwardedHost)
			}

			redirectLocation.ServeHTTP(recorder, req)

			location := recorder.Header().Get("Location")

			if test.expLocation != location {
				t.Errorf("Unexpected redirect Location: expected %+v, result: %+v", test.expLocation, location)
			}
		})
	}
}
