// Copyright 2018 github.com/xiaoenai. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
