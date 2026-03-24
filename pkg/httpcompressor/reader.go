package httpcompressor

import (
	"io"
)

// CompressorReader defines reader interface for compressed streams.
type CompressorReader interface {
	io.ReadCloser
	Reset(r io.Reader) error
}

// CompressReader adapts compressed reader to io.ReadCloser.
type CompressReader struct {
	r          io.ReadCloser
	compressor CompressorReader
}

func newCompressReader(r io.ReadCloser, compressor CompressorReader) (*CompressReader, error) {

	return &CompressReader{
		r:          r,
		compressor: compressor,
	}, nil
}

// Read reads uncompressed bytes from underlying compressor.
func (c *CompressReader) Read(p []byte) (n int, err error) {
	return c.compressor.Read(p)
}

// Close closes both compressor and original reader.
func (c *CompressReader) Close() error {
	if err := c.compressor.Close(); err != nil {
		return err
	}
	return c.r.Close()
}
