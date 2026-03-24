package httpcompressor

import (
	"bytes"
	"compress/gzip"
	"gophermart-loyalty/internal/gopherman/constant"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewGzipWriter(t *testing.T) {
	recorder := httptest.NewRecorder()
	cw := NewGzipWriter(recorder)
	if cw == nil {
		t.Fatal("NewGzipWriter() returned nil")
	}
	testData := "Hello, Gzip!"
	_, err := cw.Write([]byte(testData))
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if err := cw.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if got := recorder.Header().Get(constant.ContentEncodingHeader); got != constant.GzipEncoding {
		t.Errorf("Content-Encoding = %v, want %v", got, constant.GzipEncoding)
	}

	body := recorder.Body.Bytes()
	if len(body) < 2 || body[0] != 0x1f || body[1] != 0x8b {
		t.Error("Data is not gzipped")
	}
	gr, err := gzip.NewReader(bytes.NewReader(body))
	if err != nil {
		t.Fatalf("gzip.NewReader() error = %v", err)
	}
	defer gr.Close()

	uncompressed, err := io.ReadAll(gr)
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}

	if string(uncompressed) != testData {
		t.Errorf("Uncompressed data = %v, want %v", string(uncompressed), testData)
	}
}

func TestNewGzipReader(t *testing.T) {
	testData := "Hello, Gzip Reader!"
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	_, err := gw.Write([]byte(testData))
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
	if string(uncompressed) != testData {
		t.Errorf("Read() = %v, want %v", string(uncompressed), testData)
	}
}

func TestNewGzipReader_InvalidData(t *testing.T) {
	invalidData := []byte("not gzip data")

	_, err := NewGzipReader(io.NopCloser(bytes.NewReader(invalidData)))
	if err == nil {
		t.Error("NewGzipReader() with invalid data should return error")
	}
}

func TestGzipWriter_IntegrationWithHTTP(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cw := NewGzipWriter(w)
		defer cw.Close()
		w.Header().Set(constant.ContentTypeHeader, constant.ApplicationJSON)
		cw.Write([]byte(`{"message":"hello"}`))
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(constant.AcceptEncodingHeader, constant.GzipEncoding)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)
	if got := recorder.Header().Get(constant.ContentEncodingHeader); got != constant.GzipEncoding {
		t.Errorf("Content-Encoding = %v, want %v", got, constant.GzipEncoding)
	}
	body := recorder.Body.Bytes()
	if len(body) < 2 || body[0] != 0x1f || body[1] != 0x8b {
		t.Error("Response is not gzipped")
	}
}

func TestGzipReader_IntegrationWithHTTP(t *testing.T) {
	testData := `{"id":"test","value":123}`
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	gw.Write([]byte(testData))
	gw.Close()
	compressedData := buf.Bytes()
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(compressedData))
	req.Header.Set(constant.ContentEncodingHeader, constant.GzipEncoding)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cr, err := NewGzipReader(r.Body)
		if err != nil {
			t.Fatalf("NewGzipReader() error = %v", err)
		}
		defer cr.Close()
		uncompressed, err := io.ReadAll(cr)
		if err != nil {
			t.Fatalf("ReadAll() error = %v", err)
		}
		if string(uncompressed) != testData {
			t.Errorf("Read() = %v, want %v", string(uncompressed), testData)
		}
		w.WriteHeader(http.StatusOK)
	})
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusOK {
		t.Errorf("Status code = %v, want %v", recorder.Code, http.StatusOK)
	}
}
