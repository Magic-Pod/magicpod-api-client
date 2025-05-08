package common

import (
	"os"
	"testing"
)


func TestZipAppDir(t *testing.T) {
	dir := "testdata/compression"

	zipFile := zipAppDir(dir)
	fileInfo, err := os.Stat(zipFile)
	if err != nil {
		t.Fatalf("Failed to read zip file: %v", err)
	}

	FILE_SIZE_THRESHOLD := int64(4300)
	if fileInfo.Size() >= FILE_SIZE_THRESHOLD {
		t.Errorf("Zip file size %d is larger than original directory size %d", fileInfo.Size(), FILE_SIZE_THRESHOLD)
	}
}
