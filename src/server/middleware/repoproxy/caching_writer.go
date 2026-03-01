package repoproxy

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/goharbor/harbor/src/controller/proxy"
	"github.com/goharbor/harbor/src/lib/log"
)

type CachingResponseWriter struct {
	http.ResponseWriter
	file       *os.File
	digest     string
	written    int64
	tempPath   string
	statusCode int
}

func NewCachingResponseWriter(w http.ResponseWriter, digest string) *CachingResponseWriter {
	cacheDir := filepath.Join(os.TempDir(), "lru_blob_cache")
	os.MkdirAll(cacheDir, 0755)

	tempPath := filepath.Join(cacheDir, digest+".tmp")
	f, err := os.OpenFile(tempPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		log.Errorf("Failed to open cache temp file for %s: %v", digest, err)
		f = nil
	}

	return &CachingResponseWriter{
		ResponseWriter: w,
		file:           f,
		digest:         digest,
		tempPath:       tempPath,
		statusCode:     0,
	}
}

func (w *CachingResponseWriter) WriteHeader(statusCode int) {
	if w.statusCode == 0 {
		w.statusCode = statusCode
		w.ResponseWriter.WriteHeader(statusCode)
	}
}

func (w *CachingResponseWriter) Write(b []byte) (int, error) {
	if w.statusCode == 0 {
		w.WriteHeader(http.StatusOK)
	}
	n, err := w.ResponseWriter.Write(b)
	if w.statusCode >= 200 && w.statusCode < 300 && n > 0 && w.file != nil {
		w.file.Write(b[:n])
		w.written += int64(n)
	}
	return n, err
}

func (w *CachingResponseWriter) Close() {
	if w.file != nil {
		w.file.Close()

		if w.statusCode >= 200 && w.statusCode < 300 {
			contentLengthStr := w.Header().Get("Content-Length")
			expectedSize, _ := strconv.ParseInt(contentLengthStr, 10, 64)

			if expectedSize > 0 && w.written == expectedSize {
				finalPath := filepath.Join(filepath.Dir(w.tempPath), w.digest)
				os.Rename(w.tempPath, finalPath)
				if proxy.BlobCache != nil {
					proxy.BlobCache.Add(context.Background(), w.digest, w.written)
				}
				log.Infof("LRU Cache: Successfully cached blob %s to local disk (size: %d)", w.digest, w.written)
				return
			}
			log.Debugf("LRU Cache: Incomplete blob read across proxy. Written: %d, expected: %d", w.written, expectedSize)
		}
		os.Remove(w.tempPath)
	}
}
