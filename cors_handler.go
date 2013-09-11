package endpoint

import (
	"net/http"
	"strings"
)

const CORS_HEADER_ORIGIN = "Origin"
const CORS_HEADER_REQUEST_METHOD = "Access-Control-Request-Method"
const CORS_HEADER_REQUEST_HEADERS = "Access-Control-Request-Headers"
const CORS_HEADER_ALLOW_ORIGIN = "Access-Control-Allow-Origin"
const CORS_HEADER_ALLOW_METHODS = "Access-Control-Allow-Methods"
const CORS_HEADER_ALLOW_HEADERS = "Access-Control-Allow-Headers"

var CorsAllowedMethods = []string{"DELETE", "GET", "PATCH", "POST", "PUT"}

type corsHandler interface {
	updateHeaders(http.Header)
}

// Track information about CORS headers and our response to them.
type checkCorsHeaders struct {
	allowCorsRequest   bool
	origin               string
	corsRequestMethod  string
	corsRequestHeaders string
}

func newCheckCorsHeaders(request *http.Request) *checkCorsHeaders {
	c := &checkCorsHeaders{allowCorsRequest: false}
	c.checkCorsRequest(request)
	return c
}

// Check for a CORS request, and see if it gets a CORS response.
func (c *checkCorsHeaders) checkCorsRequest(request *http.Request) {
	// Check for incoming CORS headers.
	c.origin = request.Header.Get(CORS_HEADER_ORIGIN)
	c.corsRequestMethod = request.Header.Get(CORS_HEADER_REQUEST_METHOD)
	c.corsRequestHeaders = request.Header.Get(CORS_HEADER_REQUEST_HEADERS)

	// Check if the request should get a CORS response.
	in := false
	for _, method := range CorsAllowedMethods {
		if method == strings.ToUpper(c.corsRequestMethod) {
			in = true
			break
		}
	}
	if len(c.origin) > 0 && ((len(c.corsRequestMethod) == 0) || in) {
		c.allowCorsRequest = true
	}
}

// Add CORS headers to the response, if needed.
func (c *checkCorsHeaders) updateHeaders(headers http.Header) {
	if !c.allowCorsRequest {
		return
	}

	// Add CORS headers.
	headers.Set(CORS_HEADER_ALLOW_ORIGIN, c.origin)
	headers.Set(CORS_HEADER_ALLOW_METHODS,
		strings.Join(CorsAllowedMethods, ","))
	if len(c.corsRequestHeaders) != 0 {
		headers.Set(CORS_HEADER_ALLOW_HEADERS, c.corsRequestHeaders)
	}
}
