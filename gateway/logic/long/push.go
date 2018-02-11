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

package long

// import (
// 	"errors"
// 	"fmt"
// 	"net/url"
// 	"strconv"
// 	"sync"
// 	"sync/atomic"

// 	tp "github.com/henrylee2cn/teleport"
// 	"gitlab.xiaoenai.net/xserver/log"
// )

// var ErrIdAndRecvSeqMatch = errors.New("The number of uids and recv_seq is not equal")

// type PushReport struct {
// 	SuccCount int64   `protobuf:"varint,1,opt,name=succ_count,json=succCount,proto3" json:"succ_count,omitempty"`
// 	FailCount int64   `protobuf:"varint,2,opt,name=fail_count,json=failCount,proto3" json:"fail_count,omitempty"`
// 	FailIds   []int32 `protobuf:"varint,3,rep,packed,name=fail_ids,json=failIds" json:"fail_ids,omitempty"`
// 	mu        sync.Mutex
// }

// func InnerPush(ids []int32, recvSeqs []int64, uri string, body interface{}, bodyCodec string) (*PushReport, error) {
// 	if len(ids) != len(recvSeqs) {
// 		return nil, ErrIdAndRecvSeqMatch
// 	}
// 	u, err := url.Parse(uri)
// 	if err != nil {
// 		return nil, fmt.Errorf("Incorrect uri: %s", err.Error())
// 	}
// 	var report = new(PushReport)
// 	q := u.Query()
// 	q.Del("recv_seq")
// 	var uriPrefix string
// 	if len(q) > 0 {
// 		uriPrefix = u.Path + "?" + q.Encode() + "&recv_seq="
// 	} else {
// 		uriPrefix = u.Path + "?recv_seq="
// 	}
// 	var wg sync.WaitGroup
// 	for i, id := range ids {
// 		idStr := strconv.Itoa(int(id))
// 		sess, ok := peer.GetSession(idStr)
// 		if !ok {
// 			forcedRemoveAgent(idStr)
// 			report.AddFailIds(id)
// 			log.Warnf("Before pushing to id[%s], no connection found", idStr)
// 			continue
// 		}
// 		recvSeq := recvSeqs[i]
// 		_id := id
// 		wg.Add(1)
// 		if !tp.Go(func() {
// 			defer wg.Done()
// 			err = push(sess, uriPrefix+strconv.FormatInt(recvSeq, 10), body, bodyCodec)
// 			if err != nil {
// 				report.AddFailIds(_id)
// 				log.Warnf("When pushing to id[%d], the connection is broken: %s", _id, err.Error())
// 			} else {
// 				atomic.AddInt64(&report.SuccCount, 1)
// 			}
// 		}) {
// 			wg.Done()
// 			report.AddFailIds(_id)
// 		}
// 	}
// 	wg.Wait()
// 	report.FailCount = int64(len(ids)) - report.SuccCount
// 	return report, nil
// }

// func InnerBroadcast(uri string, body interface{}, bodyCodec string) *PushReport {
// 	var report = new(PushReport)
// 	var wg sync.WaitGroup
// 	peer.RangeSession(func(sess tp.Session) bool {
// 		sessId := sess.Id()
// 		_id, _ := strconv.Atoi(sessId)
// 		id := int32(_id)
// 		wg.Add(1)
// 		if !tp.Go(func() {
// 			defer wg.Done()
// 			if e := push(sess, uri, body, bodyCodec); e != nil {
// 				report.AddFailIds(id)
// 				atomic.AddInt64(&report.FailCount, 1)
// 				log.Debugf("When pushing to id[%s], the connection is broken: %s", sessId, e.Error())
// 			} else {
// 				atomic.AddInt64(&report.SuccCount, 1)
// 			}
// 		}) {
// 			wg.Done()
// 			report.AddFailIds(id)
// 		}
// 		return true
// 	})
// 	wg.Wait()
// 	return report
// }

// func (p *PushReport) AddFailIds(id ...int32) {
// 	p.mu.Lock()
// 	p.FailIds = append(p.FailIds, id...)
// 	p.mu.Unlock()
// }
