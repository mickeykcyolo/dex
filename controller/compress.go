package controller

import (
	"github.com/cyolo-core/pkg/pools"
	"io"
	"net/http"
	"strings"
)

// gzipWriterPool is a free-list of gzip writers.
var gzipWriterPool = pools.NewGzipWriterPool()

// gzipResponseWriter implements idiomatic compression
// for golang std http.Handlers.
type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

// Write implements io.Writer for gzipResponseWriter.
func (w gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

// compress middleware implements gzip compression for the next
// http handler.
func compress(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}
		w.Header().Set("Content-Encoding", "gzip")
		gz := gzipWriterPool.Get(w)
		next.ServeHTTP(gzipResponseWriter{Writer: gz, ResponseWriter: w}, r)
		gz.Close()
		gzipWriterPool.Put(gz)
	})
}
