package httpapi

import (
	"bytes"
	"os"
	"testing"
)

// TestOpenAPISpecMatchesDocs guarantees that the spec embedded in the binary
// (static/openapi.yaml) is identical to the documented one in docs/openapi.yaml.
// Editing only one of the two copies makes this test fail.
func TestOpenAPISpecMatchesDocs(t *testing.T) {
	// Tests run inside the package directory: backend/internal/httpapi -> repo root.
	const docsPath = "../../../docs/openapi.yaml"

	docs, err := os.ReadFile(docsPath)
	if err != nil {
		t.Fatalf("could not read %s: %v", docsPath, err)
	}

	if !bytes.Equal(docs, openapiSpec) {
		t.Fatalf("static/openapi.yaml differs from %s; copy the file to keep them in sync", docsPath)
	}
}
