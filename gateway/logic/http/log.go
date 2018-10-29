package http

import (
	"math"
	"strconv"
	"time"

	tp "github.com/henrylee2cn/teleport"
	"github.com/henrylee2cn/teleport/plugin"
	"github.com/henrylee2cn/teleport/utils"
)

var (
	printDetail       bool
	countTime         bool
	slowCometDuration time.Duration = math.MaxInt64
)

func (r *requestHandler) runlog(startTime time.Time, label *plugin.ProxyLabel, seq string, inputBody []byte, outputBody *[]byte) {
	var (
		costTimeStr string
		printFunc   = tp.Infof
		statusCode  = r.ctx.Response.StatusCode()
	)
	if countTime {
		costTime := time.Since(startTime)
		costTimeStr = costTime.String()
		if costTime >= slowCometDuration {
			costTimeStr += "(slow)"
			printFunc = tp.Warnf
		} else {
			costTimeStr += "(fast)"
		}
	} else {
		costTimeStr = "()"
	}

	if statusCode == 500 {
		printFunc = tp.Errorf
	}

	if statusCode != 200 && statusCode != 404 && statusCode != 500 {
		printFunc = tp.Warnf
	}

	printFunc("PULL<- %d %q %s %s %q RECV(%s) SEND(%s)", statusCode, label.RealIp, costTimeStr, label.Uri, seq, r.packetLogBytes(inputBody, r.ctx.Request.Header.Header(), false), r.packetLogBytes(*outputBody, r.ctx.Response.Header.Header(), r.errMsg != nil))
}

func (r *requestHandler) packetLogBytes(bodyBytes, headerBytes []byte, hasErr bool) []byte {
	var size = len(bodyBytes) + len(headerBytes)
	if hasErr {
		size += len(r.errMsg)
	}
	var b = make([]byte, 0, 128)
	b = append(b, '{')
	b = append(b, '"', 's', 'i', 'z', 'e', '"', ':')
	b = append(b, strconv.FormatUint(uint64(size), 10)...)
	if hasErr {
		b = append(b, ',', '"', 'e', 'r', 'r', 'o', 'r', '"', ':')
		if len(r.errMsg) > 0 && r.errMsg[0] == '{' {
			b = append(b, r.errMsg...)
		} else {
			b = append(b, utils.ToJsonStr(r.errMsg, false)...)
		}

	}
	if printDetail {
		b = append(b, ',', '"', 'm', 'e', 't', 'a', '"', ':')
		b = append(b, utils.ToJsonStr(headerBytes, false)...)
		if !hasErr && len(bodyBytes) > 0 {
			b = append(b, ',', '"', 'b', 'o', 'd', 'y', '"', ':')
			if bodyBytes[0] == '{' {
				b = append(b, bodyBytes...)
			} else {
				b = append(b, utils.ToJsonStr(bodyBytes, false)...)
			}
		}
	}
	b = append(b, '}')
	return b
}
