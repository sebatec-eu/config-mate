package csrf

import (
	"net/http"
)

func mitigationMiddleware(headerName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			isGet := r.Method == "GET"
			value, ok := r.Header[headerName]
			if !isGet && !ok {
				http.Error(w, "unauthorized request", http.StatusUnauthorized)
				return
			}

			if ok && len(value) > 1 {
				http.Error(w, "unauthorized request", http.StatusUnauthorized)
				return
			}

			if ok && value[0] != "1" {
				http.Error(w, "unauthorized request", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func MitigationMiddleware() func(next http.Handler) http.Handler {
	return mitigationMiddleware("X-Sebatec-Csrf-Protection")
}
