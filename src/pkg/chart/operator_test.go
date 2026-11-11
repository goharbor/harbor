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
	"helm.sh/helm/v3/pkg/chart/loader"
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

// TestGetData_HelmDecompressedLimit guards that GetData relies on helm's
// loader to bound decompression (loader.MaxDecompressedChartSize).
func TestGetData_HelmDecompressedLimit(t *testing.T) {
	chartOpr := NewOperator()

	// Only the tar header declaring an oversized entry is needed: helm rejects
	// it from the header before reading any data, so the test stays fast and
	// does not allocate the declared size.
	var buf bytes.Buffer
	gzw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gzw)
	header := &tar.Header{
		Name: "harbor/values.yaml",
		Size: loader.MaxDecompressedChartSize + 1,
		Mode: 0600,
	}
	// The entry is intentionally left incomplete, so tw is not closed.
	assert.Nil(t, tw.WriteHeader(header))
	assert.Nil(t, gzw.Close())

	_, err := chartOpr.GetData(buf.Bytes())
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "is larger than the maximum file size")
}
