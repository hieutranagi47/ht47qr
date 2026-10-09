package http

import _ "embed"

//go:embed openapi.yaml
var openAPISpec []byte

// OpenAPISpec returns the module's OpenAPI document for publication by the
// service-level documentation hub.
func OpenAPISpec() []byte { return append([]byte(nil), openAPISpec...) }
