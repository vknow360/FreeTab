package updater

import (
	"archive/zip"
	"bytes"
	"os"
	"testing"
)

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		current  string
		latest   string
		expected bool
	}{
		{"1.1", "1.2", true},
		{"v1.1", "v1.2", true},
		{"1.1", "1.1", false},
		{"v1.1", "v1.1", false},
		{"1.1", "1.1.1", true},
		{"1.1.1", "1.1", false},
		{"1.1", "2.0", true},
		{"2.0", "1.1", false},
		{"1.1", "1.0.9", false},
	}

	for _, tc := range tests {
		got := CompareVersions(tc.current, tc.latest)
		if got != tc.expected {
			t.Errorf("CompareVersions(%q, %q) = %v; expected %v", tc.current, tc.latest, got, tc.expected)
		}
	}
}

func TestFindAsset(t *testing.T) {
	rel := &GitHubRelease{
		TagName: "v1.2",
		Assets: []GitHubAsset{
			{Name: "README.md", DownloadURL: "https://example.com/readme.md"},
			{Name: "FreeTab.exe", DownloadURL: "https://example.com/FreeTab.exe"},
			{Name: "FreeTab_Portable.zip", DownloadURL: "https://example.com/FreeTab_Portable.zip"},
		},
	}

	asset, err := FindAsset(rel)
	if err != nil {
		t.Fatalf("FindAsset failed: %v", err)
	}

	if asset.Name != "FreeTab.exe" {
		t.Errorf("expected FreeTab.exe, got %s", asset.Name)
	}
}

func TestExtractExecutableFromZip(t *testing.T) {
	// Create an in-memory zip containing FreeTab.exe
	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)

	w, err := zw.Create("FreeTab.exe")
	if err != nil {
		t.Fatalf("failed to create zip entry: %v", err)
	}

	testData := []byte("MZ_TEST_BINARY_DATA")
	if _, err := w.Write(testData); err != nil {
		t.Fatalf("failed to write to zip entry: %v", err)
	}

	if err := zw.Close(); err != nil {
		t.Fatalf("failed to close zip writer: %v", err)
	}

	// Write zip to a temporary file
	tmpZip, err := os.CreateTemp("", "test-*.zip")
	if err != nil {
		t.Fatalf("failed to create temp zip: %v", err)
	}
	defer os.Remove(tmpZip.Name())

	if _, err := tmpZip.Write(buf.Bytes()); err != nil {
		t.Fatalf("failed to write temp zip: %v", err)
	}
	tmpZip.Close()

	// Extract
	extracted, err := extractExecutableFromZip(tmpZip.Name())
	if err != nil {
		t.Fatalf("extractExecutableFromZip failed: %v", err)
	}
	defer os.Remove(extracted)

	content, err := os.ReadFile(extracted)
	if err != nil {
		t.Fatalf("failed to read extracted file: %v", err)
	}

	if !bytes.Equal(content, testData) {
		t.Errorf("extracted content mismatch: got %s, expected %s", string(content), string(testData))
	}
}

func TestCleanupOldBinary(t *testing.T) {
	// Should execute without panicking even if no .old file exists
	CleanupOldBinary()
}
