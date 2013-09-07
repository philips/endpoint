// Errors used in the local Cloud Endpoints server.
package endpoint

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type RequestError interface {
	status_code() int
	rpc_error() map[string]interface{}
	rest_error() string
}

// Base type for errors that happen while processing a request.
type BaseRequestError struct {
	StatusCode int // HTTP status code number associated with this error.

	Message string // Text message explaining the error.

	// Error reason is a custom string in the Cloud Endpoints server.  When
	// possible, this should match the reason that the live server will generate,
	// based on the error's status code.  If this returns None, the error formatter
	// will attempt to generate a reason from the status code.
	Reason string

	Domain string // The string "global" by default.

	ExtraFields map[string]interface{} // Some errors have additional information. This provides a way for subclasses to provide that information.
}

func (re *BaseRequestError) status_code() int {
	return re.StatusCode
}

func (re *BaseRequestError) Error() string {
	return re.Message
}

// Format this error into a JSON response.
func (err *BaseRequestError) format_error(error_list_tag string) map[string]interface{} {
	error := map[string]interface{}{
		"domain":  err.Domain,
		"reason":  err.Reason,
		"message": err.Message,
	}
	for k, v := range err.ExtraFields {
		error[k] = v
	}
	return map[string]interface{}{
		"error": map[string]interface{}{
			error_list_tag: []map[string]interface{}{error},
			"code":         err.StatusCode,
			"message":      err.Message,
		},
	}
}

// Format this error into a response to a REST request.
func (err *BaseRequestError) rest_error() string {
	error_json := err.format_error("errors")
	rest, _ := json.MarshalIndent(error_json, "", "  ") // todo: sort keys
	return string(rest)
}

// Format this error into a response to a JSON RPC request.
func (err *BaseRequestError) rpc_error() map[string]interface{} {
	return err.format_error("data")
}

// Request rejection exception for enum values.
type EnumRejectionError struct {
	BaseRequestError
	parameter_name string   // The name of the enum parameter which had a value rejected.
	value          string   // The actual value passed in for the enum.
	allowed_values []string // List of strings allowed for the enum.
}

func NewEnumRejectionError(parameter_name, value string, allowed_values []string) *EnumRejectionError {
	return &EnumRejectionError{
		BaseRequestError: BaseRequestError{
			StatusCode: 400,
			Message:    fmt.Sprintf("Invalid string value: %s. Allowed values: %v", value, allowed_values),
			Reason:     "invalidParameter",
			ExtraFields: map[string]interface{}{
				"locationType": "parameter",
				"location":     parameter_name,
			},
		},
		parameter_name: parameter_name,
		value:          value,
		allowed_values: allowed_values,
	}
}

func (err *EnumRejectionError) Error() string {
	return err.Message
}

// Error returned when the backend SPI returns an error code.
type BackendError struct {
	BaseRequestError
	errorInfo *ErrorInfo
}

func NewBackendError(response *http.Response) *BackendError {
	// Convert backend error status to whatever the live server would return.
	error_info := get_error_info(response.StatusCode)

	var error_json map[string]interface{}
	body, _ := ioutil.ReadAll(response.Body)
	err := json.Unmarshal(body, &error_json)
	//	fmt.Printf("%s\n%s\n", string(body), err.Error())
	var message string
	if err == nil {
		_message, ok := error_json["error_message"]
		if ok {
			message, ok = _message.(string)
			if !ok {
				message = string(body)
			}
		} else {
			message = string(body)
		}
	} else {
		message = string(body)
	}

	return &BackendError{
		BaseRequestError: BaseRequestError{
			StatusCode: error_info.http_status,
			Message:    message,
			Reason:     error_info.reason,
			Domain:     error_info.domain,
		},
		errorInfo: error_info,
	}
}

func (err *BackendError) Error() string {
	return err.Message
}