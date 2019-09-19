/*
Copyright (c) grffio.

This source code is licensed under the MIT license found in the
LICENSE file in the root directory of this source tree.
*/

package httpwaterius

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDataHandle(t *testing.T) {
	testCases := []struct {
		name       string
		devices    []string
		body       string
		wantBody   string
		wantData   map[string]Data
		wantStatus int
	}{
		{
			name:       "without body",
			body:       "",
			wantBody:   "No body",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid body",
			body:       "Invalid body",
			wantBody:   "Invalid body",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "without required field: key",
			body:       `{"ch0":"1","ch1":"1","key":""}`,
			wantBody:   "Missing required field: key",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "without required field: ch0",
			body:       `{"ch0":"","ch1":"1","key":"test"}`,
			wantBody:   "Missing required fields: ch0 or ch1",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "without required field: ch1",
			body:       `{"ch0":"1","ch1":"","key":"test"}`,
			wantBody:   "Missing required fields: ch0 or ch1",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "unsupported device",
			devices:    []string{"Test"},
			body:       `{"ch0":"1","ch1":"1","key":"Unsupported"}`,
			wantBody:   "Unsupported device: Unsupported",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "all ok",
			devices:    []string{"Test"},
			body:       `{"ch0":"1","ch1":"1","key":"Test"}`,
			wantStatus: http.StatusOK,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			url := "/"
			r, err := http.NewRequest(http.MethodPost, url, strings.NewReader(tc.body))
			if err != nil {
				t.Fatal(err)
			}
			rr := httptest.NewRecorder()
			newDataHandler(tc.devices).ServeHTTP(rr, r)

			if code := rr.Code; code != tc.wantStatus {
				t.Errorf("got status: %d; want status: %d", code, tc.wantStatus)
			}
			if body := strings.Trim(rr.Body.String(), "\n"); body != tc.wantBody {
				t.Errorf("got body: %s; want body: %s", body, tc.wantBody)
			}
		})
	}
}
