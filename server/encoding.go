package server

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"strings"
)

type Encoder interface {
	Encode([]byte) ([]byte, error)
	ContentEncoding() string
}

type GzipEncoder struct{}

func (e GzipEncoder) Encode(body []byte) ([]byte, error) {
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)

	_, err := zw.Write(body)
	if err != nil {
		return nil, err
	}

	err = zw.Close()
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (e GzipEncoder) ContentEncoding() string {
	return "gzip"
}

type EncodeManager struct {
	encoders map[string]Encoder
}

func NewEncodeManager() *EncodeManager {
	manager := &EncodeManager{
		encoders: make(map[string]Encoder),
	}

	manager.RegisterEncoder("gzip", GzipEncoder{})
	return manager
}

func (manager *EncodeManager) RegisterEncoder(name string, encoder Encoder) {
	manager.encoders[name] = encoder
}

func (manager *EncodeManager) GetEncoding(acceptEncoding string) Encoder {
	if acceptEncoding == "" {
		return nil
	}

	encodings := strings.SplitSeq(acceptEncoding, ",")
	for encoding := range encodings {
		encoding = strings.TrimSpace(encoding)
		fmt.Println(encoding)

		// server supports encoding
		if encoding, ok := manager.encoders[encoding]; ok {
			return encoding
		}
	}

	return nil
}

func (manager *EncodeManager) ApplyEncoding(req *Request, res *Response) error {
	if len(res.Body) == 0 {
		return nil
	}
	acceptEncodingHeader := req.Headers["accept-encoding"]
	encoder := manager.GetEncoding(acceptEncodingHeader)

	if encoder == nil {
		return nil
	}

	encodedBody, err := encoder.Encode(res.Body)
	if err != nil {
		return err
	}

	res.Body = encodedBody
	res.ContentLength = len(encodedBody)
	res.Headers["Content-Encoding"] = encoder.ContentEncoding()

	return nil
}
