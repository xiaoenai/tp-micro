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
	"strings"

	"github.com/henrylee2cn/teleport/codec"
)

var (
	bodyCodecMapping = map[string]byte{
		"application/x-protobuf":            codec.ID_PROTOBUF,
		"application/json":                  codec.ID_JSON,
		"application/x-www-form-urlencoded": codec.ID_FORM,
		"text/plain":                        codec.ID_PLAIN,
	}
	contentTypeMapping = map[byte]string{
		codec.ID_PROTOBUF: "application/x-protobuf",
		codec.ID_JSON:     "application/json",
		codec.ID_FORM:     "application/x-www-form-urlencoded",
		codec.ID_PLAIN:    "text/plain",
	}
)

// RegBodyCodec registers a mapping of content type to body coder.
func RegBodyCodec(contentType string, codecId byte) {
	bodyCodecMapping[contentType] = codecId
	contentTypeMapping[codecId] = contentType
}

// GetBodyCodec returns the codec id from content type.
func GetBodyCodec(contentType string, defCodecId byte) byte {
	idx := strings.Index(contentType, ";")
	if idx != -1 {
		contentType = contentType[:idx]
	}
	codecId, ok := bodyCodecMapping[contentType]
	if !ok {
		return defCodecId
	}
	return codecId
}

// GetContentType returns the content type from codec id.
func GetContentType(codecId byte, defContentType string) string {
	contentType, ok := contentTypeMapping[codecId]
	if !ok {
		return defContentType
	}
	return contentType
}
