// Package traefik_plugin_redirect_ipv6 is a traefik plugin fixing the location header in a redirect response.
package traefik_plugin_redirect_ipv6 //nolint
import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strings"
)

const locationHeader string = "Location"
const (
	xRealIP        = "X-Real-Ip"
	xForwardedFor  = "X-Forwarded-For"
	cfConnectingIP = "Cf-Connecting-Ip"
)
// Rewrite definition of a replacement.
type Rewrite struct {
	Regex       string `json:"regex,omitempty" toml:"regex,omitempty" yaml:"regex,omitempty"`
	Replacement string `json:"replacement,omitempty" toml:"replacement,omitempty" yaml:"replacement,omitempty"`
}

// Config of the plugin.
type Config struct {
	Default  bool      `json:"default" toml:"default" yaml:"default"`
	Rewrites []Rewrite `json:"rewrites,omitempty" toml:"rewrites,omitempty" yaml:"rewrites,omitempty"`
}

// CreateConfig creates and initializes the plugin configuration.
func CreateConfig() *Config {
	return &Config{}
}

// RedirectIPv6Location a traefik plugin fixing the location header in a redirect response.
type RedirectIPv6Location struct {
	defaultHandling bool
	rewrites        []rewrite
	next            http.Handler
	name            string
}

type rewrite struct {
	regex       *regexp.Regexp
	replacement string
}

// New create a RedirectIPv6Location plugin instance.
func New(_ context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	rewrites := make([]rewrite, len(config.Rewrites))

	for i, rewriteConfig := range config.Rewrites {
		regexp, err := regexp.Compile(rewriteConfig.Regex)
		if err != nil {
			return nil, fmt.Errorf("error compiling regex %q: %w", rewriteConfig.Regex, err)
		}
		rewrites[i] = rewrite{
			regex:       regexp,
			replacement: rewriteConfig.Replacement,
		}
	}

	return &RedirectIPv6Location{
		defaultHandling: config.Default,
		rewrites:        rewrites,
		next:            next,
		name:            name,
	}, nil
}

func (r *RedirectIPv6Location) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	myWriter := &responseWriter{
		defaultHandlingEnabled: r.defaultHandling,
		rewrites:               r.rewrites,
		writer:                 rw,
		request:                req,
	}

	r.next.ServeHTTP(myWriter, req)
}

type responseWriter struct {
	defaultHandlingEnabled bool
	rewrites               []rewrite

	writer  http.ResponseWriter
	request *http.Request
}

func (r *responseWriter) Header() http.Header {
	return r.writer.Header()
}

func (r *responseWriter) Write(bytes []byte) (int, error) {
	return r.writer.Write(bytes)
}

func (r *responseWriter) defaultHandling(location string) string {
	locationURL, err := url.Parse(location)
	if err != nil {
		http.Error(r.writer, err.Error(), http.StatusInternalServerError)
		return ""
	}

	host := r.request.Header.Get("X-Forwarded-Host")

	if locationURL.Hostname() == host || locationURL.Host == "" {
		// path prefix
		prefix := r.request.Header.Get("X-Forwarded-Prefix")
		if strings.HasPrefix(strings.TrimPrefix(locationURL.Path, "/"), strings.TrimPrefix(prefix, "/")) {
			// it seems the service has handled the removed prefix correct so do nothing
		} else {
			oldPath := locationURL.Path
			locationURL.Path = path.Join(prefix, locationURL.Path)
			// some logging
			fmt.Println("Changed location path from ", oldPath, "to", locationURL.Path)
		}
	}

	return locationURL.String()
}

func (r *responseWriter) handleRewrites(location string) string {
	for _, rewrite := range r.rewrites {
		if (rewrite.regex.MatchString(location)) {
			locationOld := location
			location = rewrite.regex.ReplaceAllString(location, rewrite.replacement)
			// some logging
			fmt.Println("Changed location from ", locationOld, "to", location)
		}
	}

	return location
}

func  (r *responseWriter) isFromIpv6() bool {
	realIP := r.request.RemoteAddr
	forwardedIPs := strings.Split(r.request.Header.Get(xForwardedFor), ",")

	for i := len(forwardedIPs) - 1; i >= 0; i-- {
		// TODO - Check if TrimSpace is necessary
		trimmedIP := strings.TrimSpace(forwardedIPs[i])
		if trimmedIP != "" {
			realIP = trimmedIP
			break
		}
	}
	
	if realIP == "" {
		realIP = r.request.Header.Get(cfConnectingIP)
	}
	fmt.Println("realIP ", realIP)
	return strings.Contains(realIP, ":")
}



func (r *responseWriter) WriteHeader(statusCode int) {
		oldURL := rawURL(r.request)

		// rewrites
		if r.isFromIpv6() && len(r.rewrites) > 0 {
			newURL := r.handleRewrites(oldURL)
			if (oldURL != newURL) {
				r.writer.Header().Set(locationHeader, newURL)
				statusCode = 301
			}
		}

		
	// call the wrapped writer
	r.writer.WriteHeader(statusCode)
}


func rawURL(req *http.Request) string {
	scheme := "http"
	host := req.Host
	port := ""
	var uri string
	if req.RequestURI != "" {
		uri = req.RequestURI
	} else if req.URL.RawPath == "" {
		uri = req.URL.Path
	} else {
		uri = req.URL.RawPath
	}


	if req.TLS != nil {
		scheme = "https"
	}

	return strings.Join([]string{scheme, "://", host, port, uri}, "")
}
