// Copyright 2013 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package endpoint

import (
	"net/http"
	"fmt"
	"encoding/json"
)

func sendNotFoundResponse(w http.ResponseWriter, corsHandler corsHandler) string {
	if corsHandler != nil {
		corsHandler.updateHeaders(w.Header())
	}
	body := "Not Found"
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(body)))
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprint(w, body)
	return body
}

func sendErrorResponse(message string, w http.ResponseWriter, corsHandler corsHandler) string {
	bodyMap := map[string]interface{}{
		"error": map[string]interface{}{
			"message": message,
		},
	}
	bodyBytes, _ := json.Marshal(bodyMap)
	body := string(bodyBytes)
	if corsHandler != nil {
		corsHandler.updateHeaders(w.Header())
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(body)))
	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprint(w, body)
	return body
}

func sendRejectedResponse(rejectionError map[string]interface{}, w http.ResponseWriter, corsHandler corsHandler) string {
	bodyBytes, _ := json.Marshal(rejectionError)
	body := string(bodyBytes)
	if corsHandler != nil {
		corsHandler.updateHeaders(w.Header())
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(body)))
	w.WriteHeader(http.StatusBadRequest)
	fmt.Fprint(w, body)
	return body
}

func sendRedirectResponse(redirectLocation string, w http.ResponseWriter, r *http.Request, corsHandler corsHandler) string {
	if corsHandler != nil {
		corsHandler.updateHeaders(w.Header())
	}
	http.Redirect(w, r, redirectLocation, http.StatusFound)
	return ""
}
