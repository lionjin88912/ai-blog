package toolchain

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestDownloadFile(t *testing.T) {
	// Setup test server
	content := []byte("hello binary content")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(content)
	}))
	defer srv.Close()

	destDir := t.TempDir()
	destFile := filepath.Join(destDir, "testfile")

	err := DownloadFile(srv.URL+"/test.bin", destFile)
	if err != nil {
		t.Fatalf("DownloadFile failed: %v", err)
	}

	got, err := os.ReadFile(destFile)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	if string(got) != string(content) {
		t.Errorf("content mismatch: got %q, want %q", got, content)
	}
}

func TestDownloadFileHTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	destFile := filepath.Join(t.TempDir(), "testfile")
	err := DownloadFile(srv.URL+"/missing", destFile)
	if err == nil {
		t.Error("expected error for 404, got nil")
	}
}

func TestExtractZip(t *testing.T) {
	// Create a minimal zip in memory
	zipPath := filepath.Join("testdata", "test.zip")
	if _, err := os.Stat(zipPath); os.IsNotExist(err) {
		t.Skip("testdata/test.zip not found, skipping zip extract test")
	}

	destDir := t.TempDir()
	err := ExtractArchive(zipPath, destDir)
	if err != nil {
		t.Fatalf("ExtractArchive failed: %v", err)
	}
}
