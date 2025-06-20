package csrf

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newRequestWithHeader(method string, header *http.Header) *http.Request {
	req := httptest.NewRequest(method, "http://example.com/foo", nil)

	if header != nil {
		req.Header = *header
	}

	return req
}

func TestCsrfMitigationMiddleware(t *testing.T) {
	handler := mitigationMiddleware("X-Sebatec-Csrf-Protection")(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, "Hello, client")
		}))

	for _, tc := range []struct {
		expect int
		req    *http.Request
	}{
		{http.StatusOK, newRequestWithHeader("GET", nil)},
		{http.StatusUnauthorized, newRequestWithHeader("POST", nil)},
		{http.StatusOK, newRequestWithHeader("GET", &http.Header{"X-Random-Header": []string{"foobar"}})},
		{http.StatusUnauthorized, newRequestWithHeader("POST", &http.Header{"X-Random-Header": []string{"foobar"}})},
		{http.StatusUnauthorized, newRequestWithHeader("GET", &http.Header{"X-Sebatec-Csrf-Protection": []string{"abc", "0"}})},
		{http.StatusUnauthorized, newRequestWithHeader("POST", &http.Header{"X-Sebatec-Csrf-Protection": []string{"abc", "0"}})},
		{http.StatusUnauthorized, newRequestWithHeader("GET", &http.Header{"X-Sebatec-Csrf-Protection": []string{"abc"}})},
		{http.StatusUnauthorized, newRequestWithHeader("GET", &http.Header{"X-Sebatec-Csrf-Protection": []string{"0"}})},
		{http.StatusOK, newRequestWithHeader("GET", &http.Header{"X-Sebatec-Csrf-Protection": []string{"1"}})},
		{http.StatusOK, newRequestWithHeader("POST", &http.Header{"X-Sebatec-Csrf-Protection": []string{"1"}})},
	} {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, tc.req)

		resp := w.Result()

		if tc.expect != resp.StatusCode {
			t.Errorf("expected status code %v but got %v", tc.expect, resp.StatusCode)
		}
	}

}
