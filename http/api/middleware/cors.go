package middleware

import "net/http"

// Cors sets the cross-origin resource sharing HTTP headers.
func Cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := w.Header()

		header["Access-Control-Allow-Origin"] = []string{"*"}
		header["Access-Control-Allow-Headers"] = []string{"Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, accept, origin, Cache-Control, X-Requested-With"}
		header["Access-Control-Allow-Methods"] = []string{"GET, POST, HEAD"}
		header["Access-Control-Max-Age"] = []string{"600"}
		header["Vary"] = []string{"Origin, Access-Control-Request-Method, Access-Control-Request-Headers"}

		// Non CORS headers added for extra security
		header["X-Xss-Protection"] = []string{"1; mode=block"}
		header["Strict-Transport-Security"] = []string{"max-age=63072000; includeSubDomains; preload"}
		header["X-Frame-Options"] = []string{"DENY"}
		header["X-Content-Type-Options"] = []string{"nosniff"}
		// header["Content-Security-Policy"] = []string{"default-src 'self'"}
		header["X-Permitted-Cross-Domain-Policies"] = []string{"none"}
		header["Referrer-Policy"] = []string{"no-referrer"}
		header["Feature-Policy"] = []string{"microphone 'none'; camera 'none'"}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		next.ServeHTTP(w, r)
	})
}
