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

// cacheDirPerm restricts cache directory access to owner only.
const cacheDirPerm = 0700

type CachingResponseWriter struct {
	http.ResponseWriter
	file       *os.File
	digest     string
	written    int64
	tempPath   string
	statusCode int
	writeErr   bool // set on first disk write failure
}

func NewCachingResponseWriter(w http.ResponseWriter, digest string) *CachingResponseWriter {
	cacheDir := filepath.Join(os.TempDir(), "lru_blob_cache")
	os.MkdirAll(cacheDir, cacheDirPerm)

	// Use a per-request unique temp file to avoid races when multiple
	// goroutines pull the same digest concurrently.
	f, err := os.CreateTemp(cacheDir, digest+".*.tmp")
	if err != nil {
		log.Errorf("Failed to create cache temp file for %s: %v", digest, err)
		return &CachingResponseWriter{ResponseWriter: w, digest: digest}
	}
	// Restrict file permissions to owner only.
	os.Chmod(f.Name(), 0600)

	return &CachingResponseWriter{
		ResponseWriter: w,
		file:           f,
		digest:         digest,
		tempPath:       f.Name(),
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
	if w.statusCode >= 200 && w.statusCode < 300 && n > 0 && w.file != nil && !w.writeErr {
		if _, wErr := w.file.Write(b[:n]); wErr != nil {
			log.Errorf("LRU Cache: disk write error for %s, abandoning cache: %v", w.digest, wErr)
			w.writeErr = true
			w.file.Close()
			os.Remove(w.tempPath)
			w.file = nil
		} else {
			w.written += int64(n)
		}
	}
	return n, err
}

func (w *CachingResponseWriter) Close() {
	if w.file != nil {
		w.file.Close()

		if w.statusCode >= 200 && w.statusCode < 300 && !w.writeErr {
			contentLengthStr := w.Header().Get("Content-Length")
			expectedSize, _ := strconv.ParseInt(contentLengthStr, 10, 64)

			if expectedSize > 0 && w.written == expectedSize {
				finalPath := filepath.Join(filepath.Dir(w.tempPath), w.digest)
				if err := os.Rename(w.tempPath, finalPath); err != nil {
					log.Errorf("LRU Cache: failed to finalize blob %s: %v", w.digest, err)
					os.Remove(w.tempPath)
					return
				}
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
