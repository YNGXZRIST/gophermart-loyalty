package httpcompressor

import (
	"io"
)

type CompressorReader interface {
	io.ReadCloser
	Reset(r io.Reader) error
}
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
func (c *CompressReader) Read(p []byte) (n int, err error) {
	return c.compressor.Read(p)
}
func (c *CompressReader) Close() error {
	if err := c.compressor.Close(); err != nil {
		return err
	}
	return c.r.Close()
}
