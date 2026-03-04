package httpcompressor

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCompressWriter_Header(t *testing.T) {
	recorder := httptest.NewRecorder()
	gw := gzip.NewWriter(recorder)
	cw := newCompressWriter(recorder, gw, "gzip")
	cw.Header().Set("X-Test", "value")
	if got := cw.Header().Get("X-Test"); got != "value" {
		t.Errorf("Header() = %v, want %v", got, "value")
	}
}

func TestCompressWriter_Write(t *testing.T) {
	tests := []struct {
		name     string
		data     string
		encoding string
	}{
		{
			name:     "write simple data",
			data:     "Hello, World!",
			encoding: "gzip",
		},
		{
			name:     "write json data",
			data:     `{"id":"test","type":"gauge","value":123.45}`,
			encoding: "gzip",
		},
		{
			name:     "write empty data",
			data:     "",
			encoding: "gzip",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			gw := gzip.NewWriter(recorder)
			cw := newCompressWriter(recorder, gw, tt.encoding)
			n, err := cw.Write([]byte(tt.data))
			if err != nil {
				t.Fatalf("Write() error = %v", err)
			}
			if n != len(tt.data) {
				t.Errorf("Write() wrote %d bytes, want %d", n, len(tt.data))
			}
			if got := recorder.Header().Get(ContentEncodingHeader); got != tt.encoding {
				t.Errorf("Content-Encoding = %v, want %v", got, tt.encoding)
			}
			if err := cw.Close(); err != nil {
				t.Fatalf("Close() error = %v", err)
			}
			if len(tt.data) > 0 {
				body := recorder.Body.Bytes()
				if len(body) < 2 || body[0] != 0x1f || body[1] != 0x8b {
					t.Errorf("Data is not gzipped, got first bytes: %x", body[:min(2, len(body))])
				}
				gr, err := gzip.NewReader(recorder.Body)
				if err != nil {
					t.Fatalf("gzip.NewReader() error = %v", err)
				}
				defer gr.Close()

				uncompressed, err := io.ReadAll(gr)
				if err != nil {
					t.Fatalf("ReadAll() error = %v", err)
				}

				if string(uncompressed) != tt.data {
					t.Errorf("Uncompressed data = %v, want %v", string(uncompressed), tt.data)
				}
			}
		})
	}
}

func TestCompressWriter_WriteHeader(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		encoding   string
	}{
		{
			name:       "status 200",
			statusCode: http.StatusOK,
			encoding:   "gzip",
		},
		{
			name:       "status 404",
			statusCode: http.StatusNotFound,
			encoding:   "gzip",
		},
		{
			name:       "status 500",
			statusCode: http.StatusInternalServerError,
			encoding:   "gzip",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			gw := gzip.NewWriter(recorder)
			cw := newCompressWriter(recorder, gw, tt.encoding)

			cw.WriteHeader(tt.statusCode)

			if got := recorder.Code; got != tt.statusCode {
				t.Errorf("Status code = %v, want %v", got, tt.statusCode)
			}
			if got := recorder.Header().Get(ContentEncodingHeader); got != tt.encoding {
				t.Errorf("Content-Encoding = %v, want %v", got, tt.encoding)
			}
		})
	}
}

func TestCompressWriter_WriteHeader_OnlyOnce(t *testing.T) {
	recorder := httptest.NewRecorder()
	gw := gzip.NewWriter(recorder)
	cw := newCompressWriter(recorder, gw, "gzip")
	cw.WriteHeader(http.StatusOK)
	cw.WriteHeader(http.StatusNotFound)
	if got := recorder.Code; got != http.StatusOK {
		t.Errorf("Status code = %v, want %v", got, http.StatusOK)
	}
}

func TestCompressWriter_Close(t *testing.T) {
	recorder := httptest.NewRecorder()
	gw := gzip.NewWriter(recorder)
	cw := newCompressWriter(recorder, gw, "gzip")
	_, err := cw.Write([]byte("test data"))
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if err := cw.Close(); err != nil {
		t.Errorf("Close() error = %v", err)
	}
	body := recorder.Body.Bytes()
	if len(body) == 0 {
		t.Error("Body is empty after Close()")
	}
}
