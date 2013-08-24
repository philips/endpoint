// Cloud Endpoints API request-related data and functions.
package discovery

import (
	"net/http"
	"encoding/json"
	"io/ioutil"
	"fmt"
	"log"
	"strings"
	"errors"
	"bytes"
)

const _API_PREFIX = "/_ah/api/"

type JsonObject map[string]interface{}

// Simple data object representing an API request.
type ApiRequest struct {
	*http.Request

//	relative_url string
	is_batch bool
	body_json JsonObject
	request_id string
}

func newApiRequest(r *http.Request) (*ApiRequest, error) {
	ar := &ApiRequest{
		Request: r,
//		reconstruct_relative_url(r),
		is_batch: false,
	}

	if !strings.HasPrefix(ar.URL.Path, _API_PREFIX) {
		return nil, fmt.Errorf("Invalid request path: %s", ar.URL.Path)
	}
	ar.URL.Path = ar.URL.Path[len(_API_PREFIX):]

	/*if len(ar.URL.RawQuery) > 0 {
		self.parameters = cgi.parse_qs(self.query, keep_blank_values=True)
	} else {
		self.parameters = {}
	}*/

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("Problem parsing request body: %s", r.Body)
	}
	var body_json interface{}
	if len(body) > 0 {
		err := json.Unmarshal(body, &body_json)
		if err != nil {
			return nil, fmt.Errorf("Problem unmarshalling request body: %s", body)
		}
	} else {
		body_json = make(JsonObject)
	}

	ar.request_id = ""//nil

	// Check if it's a batch request.  We'll only handle single-element batch
	// requests on the dev server (and we need to handle them because that's
	// what RPC and JS calls typically show up as).  Pull the request out of the
	// list and record the fact that we're processing a batch.
	body_json_array, ok := body_json.([]interface{})
	if ok {
		switch n := len(body_json_array); n {
		case 0:
			return nil, errors.New("Batch request has zero parts")
		case 1:
		default:
			log.Printf(`Batch requests with more than 1 element aren't
				supported. Only the first element
				will be handled. Found %d elements.`, n)
		}
		log.Print("Converting batch request to single request.")
		body_json = body_json_array[0]
		ar.body_json, ok = body_json.(JsonObject)
		if !ok {
			return nil, fmt.Errorf("JSON request body must be a map: %s", body_json)
		}
		var body_bytes []byte
		body_bytes, err = json.Marshal(ar.body_json)
		ar.Body = ioutil.NopCloser(bytes.NewBuffer(body_bytes))
		if err != nil {
			return ar, err
		}
		ar.is_batch = true
	} else {
		ar.body_json, ok = body_json.(JsonObject)
		if !ok {
			return nil, fmt.Errorf("JSON request body must be a map: %s", body_json)
		}
		ar.is_batch = false
	}
	return ar, nil
}

// Reconstruct the relative URL of this request.
//
// This is based on the URL reconstruction code in Python PEP 333:
// http://www.python.org/dev/peps/pep-0333/#url-reconstruction. Rebuild the
// URL from the pieces available in the environment.
//
// Args:
//   environ: An environ dict for the request as defined in PEP-333.
//
// Returns:
//   The portion of the URL from the request after the server and port.
//func reconstruct_relative_url(r *http.Request) string {
//	/*url = urllib.quote(environ.get('SCRIPT_NAME', ''))
//	url += urllib.quote(environ.get('PATH_INFO', ''))
//	if environ.get('QUERY_STRING')
//	url += '?' + environ['QUERY_STRING']*/
//	return r.URL.RequestURI
//}

func (ar *ApiRequest) copy() *ApiRequest {
//	return &ApiRequest{
//		other.ApiRequest,
//		other.relative_url,
//	}
	new_ar, _ := newApiRequest(ar.Request)
	return new_ar
}

// Google's JsonRPC protocol creates a handler at /rpc for any Cloud
// Endpoints API, with api name, version, and method name being in the
// body of the request.
// If the request is sent to /rpc, we will treat it as JsonRPC.
// The client libraries for iOS's Objective C use RPC and not the REST
// versions of the API.
func (ar *ApiRequest) is_rpc() bool {
	return ar.URL.Path == "rpc"
}

// Check if it's a batch request.  We'll only handle single-element batch
// requests on the dev server (and we need to handle them because that's
// what RPC and JS calls typically show up as).  Pull the request out of the
// list and record the fact that we're processing a batch.
/*func is_batch(r *http.Request) bool {
	var body_json interface{}
	body, _ := ioutil.ReadAll(r.Body)
	err := json.Unmarshal(body, &body_json)
	if err != nil {
		body_json = make(map[string]interface{})
	}

	if _, ok := body_json.([]interface{}); ok {
		// FIXME: Convert batch request to single request.
		return true
	}
	return false
}*/
