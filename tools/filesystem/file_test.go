package filesystem_test

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/har4s/ohmygo/tools/filesystem"
)

func TestNewFileFromFromPath(t *testing.T) {
	testDir := createTestDir(t)
	defer os.RemoveAll(testDir)

	// missing file
	_, err := filesystem.NewFileFromPath("missing")
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	// existing file
	originalName := "image_! noext"
	normalizedNamePattern := regexp.QuoteMeta("image_noext_") + `\w{10}` + regexp.QuoteMeta(".png")
	f, err := filesystem.NewFileFromPath(filepath.Join(testDir, originalName))
	if err != nil {
		t.Fatalf("Expected nil error, got %v", err)
	}
	if f.OriginalName != originalName {
		t.Fatalf("Expected originalName %q, got %q", originalName, f.OriginalName)
	}
	if match, _ := regexp.Match(normalizedNamePattern, []byte(f.Name)); !match {
		t.Fatalf("Expected Name to match %v, got %q (%v)", normalizedNamePattern, f.Name, err)
	}
	if f.Size != 73 {
		t.Fatalf("Expected Size %v, got %v", 73, f.Size)
	}
	if _, ok := f.Reader.(*filesystem.PathReader); !ok {
		t.Fatalf("Expected Reader to be PathReader, got %v", f.Reader)
	}
}
