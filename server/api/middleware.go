package api

import (
	"net/http"
)

func (s *ApiServer) tokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token == "" {
			if err := r.ParseForm(); err == nil {
				token = r.FormValue("token")
			}
		}
		if token == "" {
			token = r.URL.Query().Get("token")
		}
		if token != s.token {
			s.logger.Infow("unauthorized", "token", token)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
