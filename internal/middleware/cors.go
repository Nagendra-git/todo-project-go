package middleware

import (
	"log"
	"net/http"
)

func CORS(allowedOrigin string) func(http.Handler) http.Handler {
	log.Println("CORS Middleware Initialized for Origin:", allowedOrigin)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Printf(
				"Method=%s Path=%s Origin=%s",
				r.Method,
				r.URL.Path,
				r.Header.Get("Origin"),
			)

			// 1. Set the main origin rules
			w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			
			// 2. Dynamic Headers: Read what the browser wants, or use a broader safe-list
			requestedHeaders := r.Header.Get("Access-Control-Request-Headers")
			if requestedHeaders != "" {
				w.Header().Set("Access-Control-Allow-Headers", requestedHeaders)
			} else {
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Accept, Authorization, Origin")
			}

			// Allow credentials if your React app uses cookies/sessions (optional)
			// w.Header().Set("Access-Control-Allow-Credentials", "true")

			// 3. Handle preflight requests immediately
			if r.Method == http.MethodOptions {
				log.Println("Preflight OPTIONS request successfully intercepted and handled")
				w.WriteHeader(http.StatusNoContent) // 204 No Content is standard for OPTIONS
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
