package provider

import (
	"os"
	"path/filepath"
	"testing"
)

// writeAccConfig persists a Terraform configuration string to a local file within the 'tf_files' directory.
// This is primarily used during acceptance tests to provide a traceable record of the HCL configurations
// generated for each test step, aiding in manual debugging and verification of the provider's behavior.

func writeAccConfig(t *testing.T, name string, cfg string) {
	t.Helper()

	dir := filepath.Join("tf_files")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	file := filepath.Join(dir, name)
	if err := os.WriteFile(file, []byte(cfg), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	t.Logf("wrote acceptance config: %s", file)
}
