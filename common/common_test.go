package common

import (
	"archive/zip"
	"os"
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
