package util

import (
	"bytes"
	"compress/gzip"
	"sync"
)

var (
	gzipPool = sync.Pool{
		New: func() interface{} {
			return gzip.NewWriter(nil)
		},
	}

	bufferPool = sync.Pool{
		New: func() interface{} {
			return bytes.NewBuffer(nil)
		},
	}
)

func AcquireGzip() *gzip.Writer {
	return gzipPool.Get().(*gzip.Writer)
}

func ReleaseGzip(s *gzip.Writer) {
	gzipPool.Put(s)
}

func AcquireBuffer() *bytes.Buffer {
	return bufferPool.Get().(*bytes.Buffer)
}

func ReleaseBuffer(s *bytes.Buffer) {
	bufferPool.Put(s)
}
