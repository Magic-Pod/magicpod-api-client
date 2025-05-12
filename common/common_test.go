package common

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"testing"
)

// Verify that the zip file is created correctly and compressed to the expected size
func TestZipAppDirCompressed(t *testing.T) {
	dir := "testdata/compression"

	// Action: Create a zip file from the directory
	zipFile := zipAppDir(dir)
	fileInfo, err := os.Stat(zipFile)
	if err != nil {
		t.Fatalf("Failed to read zip file: %v", err)
	}

	// Assert: Check the size of the zip file
	const FILE_SIZE_THRESHOLD = int64(4400)
	if fileInfo.Size() >= FILE_SIZE_THRESHOLD {
		t.Errorf("Zip file size %d is larger than the threshold %d", fileInfo.Size(), FILE_SIZE_THRESHOLD)
	}

	// Assert: Check if the zip file can be opened
	zip, err := zip.OpenReader(zipFile)
	if err != nil {
		t.Errorf("Failed to open zip file: %v", err)
	}
	defer zip.Close()
}

// Verify that the symlink in the zip file points to the correct target
func TestZipAppDirSymlink(t *testing.T) {
	dirPath := "testdata/symlink"

	// Action: Create a zip file from the directory
	zipFile := zipAppDir(dirPath)
	reader, err := zip.OpenReader(zipFile)
	if err != nil {
		t.Fatalf("Failed to open zip file: %v", err)
	}
	defer reader.Close()

	for _, file := range reader.File {
		if !isSymlink(file.FileInfo()) {
			continue
		}

		rc, err := file.Open()
		if err != nil {
			t.Fatalf("Failed to open file in zip '%s': %v", file.Name, err)
		}
		defer rc.Close()

		linkTargetBytes, err := io.ReadAll(rc)
		if err != nil {
			t.Fatalf("Failed to read symlink target from zip '%s': %v", file.Name, err)
		}

		zippedLinkPath := string(linkTargetBytes)

		originalFilePath := filepath.Join(filepath.Dir(dirPath), file.Name)
		originalFile, err := os.Open(originalFilePath)
		if err != nil {
			t.Fatalf("Failed to open original symlink: %v", err)
		}
		defer originalFile.Close()

		originalLinkPath, err := os.Readlink(originalFilePath)
		if err != nil {
			t.Fatalf("Failed to read original symlink: %v", err)
		}

		// Assert: Check if the symlink in the zip file points to the correct target
		if zippedLinkPath != originalLinkPath {
			t.Errorf("Original symlink and zipped symlink do not match: %s != %s", originalLinkPath, zippedLinkPath)
		}
	}
}
