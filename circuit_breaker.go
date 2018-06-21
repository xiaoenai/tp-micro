// Copyright 2018 HenryLee. All Rights Reserved.
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

package micro

import (
	"sync"
	"time"

	"github.com/henrylee2cn/goutil"
	tp "github.com/henrylee2cn/teleport"
	sess "github.com/henrylee2cn/tp-ext/mod-cliSession"
)

const (
	// Statistical interval second
	intervalSecond = 10
	// The default failure rate threshold
	defaultErrorPercentage = 50
	// The default period of one-cycle break in milliseconds
	defaultBreakDuration = 5000 * time.Millisecond

	// circuitBreaker status

	closedStatus   = 0
	halfOpenStatus = 1
	openStatus     = 2
)

type (
	circuitBreaker struct {
		linker          Linker
		newSessionFunc  func(addr string) *cliSession
		sessLib         goutil.Map
		closeCh         chan struct{}
		enableBreak     bool
		errorPercentage float64
		breakDuration   time.Duration
	}
	cliSession struct {
		addr           string
		status         int8 // 0:Closed, 1:Half-Open, 2:Open
		succCount      [intervalSecond]int64
		failCount      [intervalSecond]int64
		cursor         int
		halfOpenTimer  *time.Timer
		halfOpenTesing bool
		rwmu           sync.RWMutex
		circuitBreaker *circuitBreaker
		*sess.CliSession
	}
)

func newCircuitBreaker(
	enableBreak bool,
	errorPercentage int,
	breakDuration time.Duration,
	linker Linker,
	newFn func(string) *sess.CliSession,
) *circuitBreaker {
	c := &circuitBreaker{
		linker:          linker,
		sessLib:         goutil.AtomicMap(),
		enableBreak:     enableBreak,
		errorPercentage: float64(errorPercentage),
		breakDuration:   breakDuration,
		closeCh:         make(chan struct{}),
	}
	c.newSessionFunc = func(addr string) *cliSession {
		return &cliSession{
			addr:           addr,
			CliSession:     newFn(addr),
			status:         closedStatus,
			circuitBreaker: c,
		}
	}
	return c
}

func (c *circuitBreaker) start() {
	go c.watchOffline()
	if c.enableBreak {
		go c.work()
	}
}

func (c *circuitBreaker) selectSession(uri string) (*cliSession, *tp.Rerror) {
	var (
		uriPath = getUriPath(uri)
		addr    string
		s       *cliSession
		exclude map[string]struct{}
		cnt     = c.linker.Len(uriPath)
		rerr    = NotFoundService
	)
	for i := cnt; i > 0; i-- {
		addr, rerr = c.linker.Select(uriPath, exclude)
		if rerr != nil {
			return nil, rerr
		}
		_s, ok := c.sessLib.Load(addr)
		if !ok {
			s = c.newSessionFunc(addr)
			c.sessLib.Store(addr, s)
			return s, nil
		}
		s = _s.(*cliSession)
		// circuit breaker check
		if !c.enableBreak || s.check() {
			return s, nil
		}
		if exclude == nil {
			exclude = make(map[string]struct{}, cnt)
		}
		exclude[addr] = struct{}{}
	}
	return s, rerr
}

func (c *circuitBreaker) work() {
	var (
		test                 = time.NewTicker(time.Second)
		state                = time.NewTicker(time.Second * intervalSecond)
		succTotal, failTotal int64
	)
	for {
		select {
		case <-test.C:
			c.sessLib.Range(func(_, _s interface{}) bool {
				s := _s.(*cliSession)
				s.rwmu.Lock()
				s.cursor++
				if s.cursor >= intervalSecond {
					s.cursor = 0
				}
				s.succCount[s.cursor] = 0
				s.failCount[s.cursor] = 0
				s.rwmu.Unlock()
				return true
			})

		case <-state.C:
			c.sessLib.Range(func(addr, _s interface{}) bool {
				s := _s.(*cliSession)
				s.rwmu.Lock()
				defer s.rwmu.Unlock()
				if s.status != closedStatus {
					return true
				}
				succTotal, failTotal = 0, 0
				for _, a := range s.failCount {
					failTotal += a
				}
				for _, a := range s.succCount {
					succTotal += a
				}
				if failTotal > 0 &&
					(float64(failTotal)/float64(failTotal+succTotal))*100 > c.errorPercentage {
					s.toOpenLocked()
					return true
				}
				s.cursor++
				if s.cursor >= intervalSecond {
					s.cursor = 0
				}
				s.succCount[s.cursor] = 0
				s.failCount[s.cursor] = 0
				return true
			})
		case <-c.closeCh:
			test.Stop()
			state.Stop()
		}
	}
}

func (c *circuitBreaker) close() {
	close(c.closeCh)
	c.linker.Close()
}

func (c *circuitBreaker) watchOffline() {
	ch := c.linker.WatchOffline()
	for addr := range <-ch {
		_s, ok := c.sessLib.Load(addr)
		if !ok {
			continue
		}
		c.sessLib.Delete(addr)
		s := _s.(*cliSession)
		if c.enableBreak && s.halfOpenTimer != nil {
			s.halfOpenTimer.Stop()
		}
		tp.Go(s.Close)
	}
}

func (s *cliSession) toOpenLocked() {
	s.status = openStatus
	s.succCount = [intervalSecond]int64{}
	s.failCount = [intervalSecond]int64{}
	s.cursor = 0
	if s.halfOpenTimer == nil {
		after := time.AfterFunc(s.circuitBreaker.breakDuration, func() {
			s.rwmu.Lock()
			s.status = halfOpenStatus
			s.halfOpenTesing = false
			s.rwmu.Unlock()
		})
		s.halfOpenTimer = after
	} else {
		s.halfOpenTimer.Reset(s.circuitBreaker.breakDuration)
	}
}

func (s *cliSession) check() bool {
	s.rwmu.RLock()
	switch s.status {
	case openStatus:
		s.rwmu.RUnlock()
		return false
	case closedStatus:
		s.rwmu.RUnlock()
		return true
	case halfOpenStatus:
		if s.halfOpenTesing {
			s.rwmu.RUnlock()
			return false
		}
		s.rwmu.RUnlock()
		s.rwmu.Lock()
		s.halfOpenTesing = true
		s.rwmu.Unlock()
		return true
	default:
		s.rwmu.RUnlock()
		return false
	}
}

func (s *cliSession) feedback(healthy bool) {
	if !s.circuitBreaker.enableBreak {
		return
	}
	s.rwmu.Lock()
	defer s.rwmu.Unlock()
	switch s.status {
	case closedStatus:
		if healthy {
			s.succCount[s.cursor]++
		} else {
			s.failCount[s.cursor]++
		}
	case halfOpenStatus:
		if healthy {
			s.status = closedStatus
		} else {
			s.toOpenLocked()
		}
	}
}
