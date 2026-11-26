package middleware

import (
	"net/http"
	"strconv"
)
type CORSConfig struct {
	AllowedOrigins map[string]bool
	AllowedMethods string
	AllowedHeaders string
	AllowedCredentials bool

}


func CorsMiddleware(cfg CORSConfig, next http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
   	origin := r.Header.Get("Origin")
    allowed := cfg.AllowedOrigins[origin]

    if origin != ""  && allowed {
     	w.Header().Set("Access-Control-Allow-Origin", origin)
      w.Header().Add("Vary", "Origin")

      w.Header().Set("Access-Control-Allow-Credentials", strconv.FormatBool(cfg.AllowedCredentials))
      w.Header().Set("Access-Control-Allow-Methods", cfg.AllowedMethods)
      w.Header().Set("Access-Control-Allow-Headers", cfg.AllowedHeaders)
    }

      // Если это preflight-запрос от браузера, просто отдаем статус 200 OK
    if r.Method == http.MethodOptions {
      w.WriteHeader(http.StatusOK)
      return
    }

      // Передаем управление дальше
    next.ServeHTTP(w, r)
  })
}
