package chart

import (
	"archive/tar"
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
