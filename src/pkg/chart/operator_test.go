package chart

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func addFileToTar(tarWriter *tar.Writer, filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return err
	}

	header, err := tar.FileInfoHeader(info, info.Name())
	if err != nil {
		return err
	}

	header.Name = strings.TrimPrefix(filename, "testdata/")

	if err = tarWriter.WriteHeader(header); err != nil {
		return err
	}

	_, err = io.Copy(tarWriter, file)
	return err
}

// loadChart archive and compress the chart folder.
func loadChart(path string) ([]byte, error) {
	compressedFileName := path + ".tar.gz"
	chartTar, err := os.Create(compressedFileName)
	if err != nil {
		return nil, err
	}

	defer func() {
		os.Remove(compressedFileName)
	}()

	gzipWritter := gzip.NewWriter(chartTar)
	tarWriter := tar.NewWriter(gzipWritter)

	if err = filepath.WalkDir(path, func(fpath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if path == fpath {
			return nil
		}

		if !d.IsDir() {
			if err = addFileToTar(tarWriter, fpath); err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		return nil, err
	}
	// flush and close writer
	tarWriter.Close()
	gzipWritter.Close()
	chartTar.Close()

	return os.ReadFile(compressedFileName)
}

func TestGetChartDetails(t *testing.T) {
	chartOpr := NewOperator()

	{
		// test schema v1 chart
		chartV1, err := loadChart("testdata/harbor-schema1")
		if err != nil {
			t.Fatalf("load chart error: %s", err)
		}

		details, err := chartOpr.GetDetails(chartV1)
		if err != nil {
			t.Fatalf("get chart details error: %s", err)
		}

		assert.Equal(t, 2, len(details.Dependencies))
		assert.Equal(t, "postgresql", details.Dependencies[0].Name)
		assert.Equal(t, "redis", details.Dependencies[1].Name)
		// the length of files should be greater than 0
		assert.Greater(t, len(details.Files), 0)
	}

	{
		// test schema v2 chart
		chartV2, err := loadChart("testdata/harbor-schema2")
		if err != nil {
			t.Fatalf("load chart error: %s", err)
		}

		details, err := chartOpr.GetDetails(chartV2)
		if err != nil {
			t.Fatalf("get chart details error: %s", err)
		}

		assert.Equal(t, 2, len(details.Dependencies))
		assert.Equal(t, "postgresql", details.Dependencies[0].Name)
		assert.Equal(t, "redis", details.Dependencies[1].Name)
		// the length of files should be greater than 0
		assert.Greater(t, len(details.Files), 0)
	}
}

func TestGetData_ExceedDecompressedLimit(t *testing.T) {
	chartOpr := NewOperator()

	// Create a tar.gz with a file of size maxDecompressedChartBytes + 1
	var buf bytes.Buffer
	gzw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gzw)

	header := &tar.Header{
		Name: "harbor/values.yaml",
		Size: int64(maxDecompressedChartBytes + 1),
		Mode: 0600,
	}
	err := tw.WriteHeader(header)
	assert.Nil(t, err)

	// Write zero bytes to satisfy the size in header (gzip handles zeros very efficiently)
	chunkSize := 1024 * 1024
	zeroChunk := make([]byte, chunkSize)
	remaining := maxDecompressedChartBytes + 1
	for remaining > 0 {
		toWrite := chunkSize
		if remaining < chunkSize {
			toWrite = remaining
		}
		_, err = tw.Write(zeroChunk[:toWrite])
		assert.Nil(t, err)
		remaining -= toWrite
	}

	tw.Close()
	gzw.Close()

	_, err = chartOpr.GetData(buf.Bytes())
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "is larger than the maximum file size")
}
