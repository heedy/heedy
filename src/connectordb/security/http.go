package security

import "net/http"

/** SecurityHeaderHandler provdies a wrapper function for an http.Handler that
sets several security headers for all sessions passing through

**/
func SecurityHeaderHandler(h http.Handler) http.Handler {

	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {

		// See the OWASP security project for these headers:
		// https://www.owasp.org/index.php/List_of_useful_HTTP_headers

		// Don't allow our site to be embedded in another
		writer.Header().Set("X-Frame-Options", "deny")

		// Enable the client side XSS filter
		writer.Header().Set("X-XSS-Protection", "1; mode=block")

		// Disable content sniffing which could lead to improperly executed
		// scripts or such from malicious user uploads
		writer.Header().Set("X-Content-Type-Options", "nosniff")

		h.ServeHTTP(writer, request)
	})
}
