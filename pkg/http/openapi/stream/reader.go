package openapistream

import (
	"bufio"
	"bytes"
	"context"
	"io"

	"github.com/go-openapi/runtime"
	"github.com/sdinsure/agent/pkg/logger"
)

type ClientResponseSinkCloser interface {
	DefaultResponse() interface{}
	SinkResponse(interface{}) error
	Close() error
}

func newStreamResponseReader(ctx context.Context, log logger.Logger, sinkCloser ClientResponseSinkCloser, orig runtime.ClientResponseReader) *streamResponseReader {
	return &streamResponseReader{
		ctx:        ctx,
		log:        log,
		orig:       orig,
		sinkCloser: sinkCloser,
	}
}

func StreamReceiveClientOption(ctx context.Context, log logger.Logger, sinkCloser ClientResponseSinkCloser) func(*runtime.ClientOperation) {
	return func(co *runtime.ClientOperation) {
		co.Reader = newStreamResponseReader(ctx, log, sinkCloser, co.Reader)
	}
}

type streamResponseReader struct {
	ctx        context.Context
	log        logger.Logger
	orig       runtime.ClientResponseReader
	sinkCloser ClientResponseSinkCloser
}

var (
	_ runtime.ClientResponseReader = &streamResponseReader{}
)

// ReadResponse implements runtime.ClientResponseReader interface
// this call is blocked until the stream is done, so this func should be run with a goroutine
func (s *streamResponseReader) ReadResponse(cr runtime.ClientResponse, cm runtime.Consumer) (interface{}, error) {
	if cr.Code() != 200 {
		// received an invalid code, forward to default readresponse and close current handling process
		s.sinkCloser.Close()
		return s.orig.ReadResponse(cr, cm)
	}

	const internalBufferSize = 10 * 1000 * 1000 /*10M*/

	buffer := make([]byte, internalBufferSize)
	defer cr.Body().Close()

	scanner := bufio.NewScanner(cr.Body())
	scanner.Buffer(buffer, internalBufferSize)
	for scanner.Scan() {
		// each response is delimitered by newline
		txt := scanner.Text()
		s.log.Infox(s.ctx, "streamreader: read len:%d\n", len(txt))

		chunked := newStaticBytesClientResponse([]byte(txt))
		resp, err := s.orig.ReadResponse(chunked, cm)
		if err != nil {
			return nil, err
		}
		select {
		case <-s.ctx.Done():
			s.log.Infox(s.ctx, "streamreader ctx is done, breaking loops\n")
			break
		default:
			if err := s.sinkCloser.SinkResponse(resp); err != nil {
				return nil, err
			}
		}
	}
	s.log.Infox(s.ctx, "streamreader: scan is done\n")
	s.sinkCloser.Close()

	// NOTE(hsiny): this should never reached until the stream is ended
	return s.sinkCloser.DefaultResponse(), nil
}

type staticBytesClientResponse struct {
	buf []byte
}

func newStaticBytesClientResponse(s []byte) staticBytesClientResponse {
	return staticBytesClientResponse{buf: s}
}

func (s staticBytesClientResponse) Code() int {
	return 200
}

func (s staticBytesClientResponse) Message() string {
	// TODO(hsiny): handle message
	return ""
}

func (s staticBytesClientResponse) GetHeader(_ string) string {
	// TODO(hsiny): handle Header
	return ""
}

func (s staticBytesClientResponse) GetHeaders(_ string) []string {
	// TODO(hsiny): handle Header
	return nil
}

func (s staticBytesClientResponse) Body() io.ReadCloser {
	return io.NopCloser(bytes.NewReader(s.buf))
}
