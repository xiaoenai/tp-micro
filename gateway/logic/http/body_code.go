package http

import (
	"bytes"

	"github.com/henrylee2cn/goutil"
	"github.com/henrylee2cn/teleport/codec"
)

var bodyCodecMapping = map[string]byte{
	"application/x-protobuf": codec.ID_PROTOBUF,
	"application/json":       codec.ID_JSON,
	"text/plain":             codec.ID_STRING,
}

// RegBodyCodec registers a mapping of content type to body coder
func RegBodyCodec(contentType string, codecId byte) {
	bodyCodecMapping[contentType] = codecId
}

// GetBodyCodec returns the codec id from content type.
func GetBodyCodec(contentTypeBytes []byte) byte {
	idx := bytes.Index(contentTypeBytes, []byte{';'})
	if idx != -1 {
		contentTypeBytes = contentTypeBytes[:idx]
	}
	contentType := goutil.BytesToString(contentTypeBytes)
	codecId, ok := bodyCodecMapping[contentType]
	if !ok {
		return codec.ID_STRING
	}
	return codecId
}
