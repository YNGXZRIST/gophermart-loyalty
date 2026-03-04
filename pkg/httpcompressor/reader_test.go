package httpcompressor

import (
	"bytes"
	"compress/gzip"
	"io"
	"strings"
	"testing"
)

func TestCompressReader_Read(t *testing.T) {
	tests := []struct {
		name string
		data string
	}{
		{
			name: "read simple data",
			data: "Hello, World!",
		},
		{
			name: "read json data",
			data: `{"id":"test","type":"gauge","value":123.45}`,
		},
		{
			name: "read large data",
			data: strings.Repeat("A", 10000),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			gw := gzip.NewWriter(&buf)
			_, err := gw.Write([]byte(tt.data))
			if err != nil {
				t.Fatalf("gzip.Write() error = %v", err)
			}
			if err := gw.Close(); err != nil {
				t.Fatalf("gzip.Close() error = %v", err)
			}
			cr, err := NewGzipReader(io.NopCloser(bytes.NewReader(buf.Bytes())))
			if err != nil {
				t.Fatalf("NewGzipReader() error = %v", err)
			}
			defer cr.Close()
			uncompressed, err := io.ReadAll(cr)
			if err != nil {
				t.Fatalf("ReadAll() error = %v", err)
			}
			if string(uncompressed) != tt.data {
				t.Errorf("Read() = %v, want %v", string(uncompressed), tt.data)
			}
		})
	}
}

func TestCompressReader_Close(t *testing.T) {
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	_, err := gw.Write([]byte("test data"))
	if err != nil {
		t.Fatalf("gzip.Write() error = %v", err)
	}
	if err := gw.Close(); err != nil {
		t.Fatalf("gzip.Close() error = %v", err)
	}
	cr, err := NewGzipReader(io.NopCloser(bytes.NewReader(buf.Bytes())))
	if err != nil {
		t.Fatalf("NewGzipReader() error = %v", err)
	}
	if err := cr.Close(); err != nil {
		t.Errorf("Close() error = %v", err)
	}
	_, err = cr.Read(make([]byte, 10))
	if err == nil {
		t.Error("Read() after Close() should return error")
	}
}

func TestCompressReader_EmptyData(t *testing.T) {
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	if err := gw.Close(); err != nil {
		t.Fatalf("gzip.Close() error = %v", err)
	}
	cr, err := NewGzipReader(io.NopCloser(bytes.NewReader(buf.Bytes())))
	if err != nil {
		t.Fatalf("NewGzipReader() error = %v", err)
	}
	defer cr.Close()
	data, err := io.ReadAll(cr)
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}

	if len(data) != 0 {
		t.Errorf("Read() returned %d bytes, want 0", len(data))
	}
}

func TestCompressReader_InvalidGzipData(t *testing.T) {
	invalidData := []byte("this is not gzip data")
	_, err := gzip.NewReader(bytes.NewReader(invalidData))
	if err == nil {
		t.Error("gzip.NewReader() with invalid data should return error")
	}
}
