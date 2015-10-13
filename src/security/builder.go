package security

import "net/http"

type SecurityBuilder struct {
	http.Handler
}

/**
	Constructs a new Security Building appliance for wrapping
	HTTP Handlers with safety/logging mechanisms.
**/
func NewSecurityBuilder(h http.Handler) SecurityBuilder {
	return SecurityBuilder{h}
}

/**
Hides 500 errors from the user, instead returning a UUID and logging the real
data on the console.
**/
func (s SecurityBuilder) Hide500Errors() SecurityBuilder {
	return SecurityBuilder{FiveHundredHandler(s)}
}

/**
Includes headers to do frame busting and prevent certain kinds of
user induced attacks.
**/
func (s SecurityBuilder) IncludeSecureHeaders() SecurityBuilder {
	return SecurityBuilder{SecurityHeaderHandler(s)}
}

/**
Logs everything that passes through your gateway, mainly a debugging tool.
**/
func (s SecurityBuilder) LogEverything() SecurityBuilder {
	return SecurityBuilder{LoggingHandler(s)}
}

func (s SecurityBuilder) Build() http.Handler {
	return s
}
