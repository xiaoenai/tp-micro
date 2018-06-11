package agent

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"
	"unsafe"

	"github.com/henrylee2cn/goutil"
	"github.com/henrylee2cn/goutil/coarsetime"
	tp "github.com/henrylee2cn/teleport"
	"github.com/henrylee2cn/teleport/plugin"
	"github.com/xiaoenai/ants/gateway/client"
	"github.com/xiaoenai/ants/gateway/logic"
	"github.com/xiaoenai/ants/gateway/logic/hosts"
	"github.com/xiaoenai/ants/gateway/logic/socket"
	"github.com/xiaoenai/ants/gateway/types"
	"github.com/xiaoenai/ants/model/redis"
)

var (
	// agentKeyPrefix agent key prefix in redis.
	agentKeyPrefix = "AGENT"
	// agentChannel agent state of the subscription channel
	agentChannel = "AGENT_Q"
	// agentLife agent max life time
	agentLife = time.Hour * 24 * 3
)

var kickUri string
var globalHandler *agentHandler

// Init initializes agent packet.
func Init(redisWithLargeMemory *redis.Client, redisWithPublishCmd *redis.Client, namespace ...string) error {
	agentKeyPrefix += ":" + logic.ApiVersion()
	agentChannel += ":" + logic.ApiVersion()
	if len(namespace) > 0 && namespace[0] != "" {
		agentKeyPrefix = namespace[0] + "@" + agentKeyPrefix
		agentChannel = namespace[0] + "@" + agentChannel
	}
	kickUri = "/gw/" + logic.ApiVersion() + "/socket_kick"
	globalHandler = new(agentHandler)
	globalHandler.redisWithLargeMemory = redisWithLargeMemory
	globalHandler.redisWithPublishCmd = redisWithPublishCmd
	globalHandler.module = redis.NewModule(agentKeyPrefix)
	return nil
}

// GetSocketHooks returns the custom agent types.SocketHooks interface.
func GetSocketHooks() types.SocketHooks {
	return globalHandler
}

type agentHandler struct {
	peer                 tp.Peer
	module               *redis.Module
	redisWithLargeMemory *redis.Client
	redisWithPublishCmd  *redis.Client
}

var (
	rerrNotOnline = tp.NewRerror(404, "Not Found", "User is not online")
)

func newSalt(m goutil.Map) uint64 {
	salt := uint64(uintptr(unsafe.Pointer(&m)))
	m.Store("agent_salt", salt)
	return salt
}

func getSalt(m goutil.Map) (uint64, bool) {
	v, ok := m.Load("agent_salt")
	if ok {
		salt, ok := v.(uint64)
		return salt, ok
	}
	return 0, false
}

var rerrServerError = tp.NewRerror(500, "System is busy, please try again later", "Agent Error")

func newServerRerror(detail string) *tp.Rerror {
	return rerrServerError.Copy().SetDetail(detail)
}

func (*agentHandler) GetSession(peer tp.Peer, sessionId string) (tp.Session, *tp.Rerror) {
	sess, ok := peer.GetSession(sessionId)
	if ok {
		return sess, nil
	}
	enforceKickOffline(sessionId, true)
	return nil, rerrNotOnline
}

func (*agentHandler) PreWritePush(tp.WriteCtx) *tp.Rerror {
	return nil
}

func (h *agentHandler) OnLogon(sess plugin.AuthSession, accessToken types.AccessToken) *tp.Rerror {
	sessionId := accessToken.SessionId()
	// check or remove old session
	_, rerr := kickOffline(sessionId, true)
	if rerr != nil {
		return rerr
	}

	// logon new agent
	_, innerIp := hosts.SocketAddress()
	a := &Agent{
		SessionId: sessionId,
		InnerGw:   innerIp,
		OnlineAt:  coarsetime.CeilingTimeNow().Unix(),
		Salt:      newSalt(sess.Swap()),
	}
	key := h.module.Key(sessionId)
	var err error
	lockErr := h.redisWithLargeMemory.LockCallback("lock_"+key, func() {
		data, _ := json.Marshal(a)
		err = h.redisWithLargeMemory.Set(key, data, agentLife).Err()
		if err == nil {
			h.redisWithPublishCmd.Publish(agentChannel, createAgentNews(sessionId, EVENT_ONLINE))
		}
	})
	if lockErr != nil {
		return newServerRerror(lockErr.Error())
	}
	if err != nil {
		return newServerRerror(err.Error())
	}
	// logon new session
	sess.SetId(sessionId)
	return nil
}

func (h *agentHandler) OnLogoff(sess tp.BaseSession) *tp.Rerror {
	salt, ok := getSalt(sess.Swap())
	if !ok {
		return nil
	}
	sessionId := sess.Id()
	_, innerGw := hosts.SocketAddress()
	var err error
	key := h.module.Key(sessionId)
	lockErr := h.redisWithLargeMemory.LockCallback(
		"lock_"+key,
		func() {
			var agentBytes []byte
			if agentBytes, err = h.redisWithLargeMemory.Get(key).Bytes(); err != nil {
				if redis.IsRedisNil(err) {
					err = nil
				}
				return
			}
			var a = new(Agent)
			if json.Unmarshal(agentBytes, a) == nil {
				if a.Salt != salt {
					return
				}
				if a.InnerGw != innerGw {
					return
				}
			}
			h.redisWithPublishCmd.Publish(agentChannel, createAgentNews(sessionId, EVENT_OFFLINE))
			err = h.redisWithLargeMemory.Del(key).Err()
		},
	)
	if lockErr != nil {
		return newServerRerror(lockErr.Error())
	}
	if err != nil {
		return newServerRerror(err.Error())
	}
	return nil
}

// EnforceKickOffline enforches kick the user offline.
func EnforceKickOffline(sessionId string) *tp.Rerror {
	return enforceKickOffline(sessionId, false)
}

// enforceKickOffline enforches kick the user offline.
func enforceKickOffline(sessionId string, checkLocal bool) *tp.Rerror {
	succ, rerr := kickOffline(sessionId, checkLocal)
	if succ || rerr != nil {
		return rerr
	}
	// enforce remove agent
	var (
		err error
		key = globalHandler.module.Key(sessionId)
	)
	lockErr := globalHandler.redisWithLargeMemory.LockCallback(
		"lock_"+key,
		func() {
			var agentBytes []byte
			if agentBytes, err = globalHandler.redisWithLargeMemory.Get(key).Bytes(); err != nil {
				if redis.IsRedisNil(err) {
					err = nil
				}
				return
			}
			var a = new(Agent)
			if json.Unmarshal(agentBytes, a) == nil {
				_, innerGw := hosts.SocketAddress()
				if a.InnerGw != innerGw {
					return
				}
			}
			globalHandler.redisWithPublishCmd.Publish(agentChannel, createAgentNews(sessionId, EVENT_OFFLINE))
			err = globalHandler.redisWithLargeMemory.Del(key).Err()
		},
	)
	if lockErr != nil {
		return newServerRerror(lockErr.Error())
	}
	if err != nil {
		return newServerRerror(err.Error())
	}
	return nil
}

// kickOffline kicks the user offline.
func kickOffline(sessionId string, checkLocal bool) (succ bool, rerr *tp.Rerror) {
	if checkLocal {
		// Try to delete the session from the local gateway.
		existed, _ := socket.Kick(sessionId)
		if existed {
			return true, nil
		}
	}
	// Find the agent of the sessionId.
	agent, rerr := GetAgent(sessionId)
	if rerr != nil {
		return false, rerr
	}
	if agent.IsOffline {
		return true, nil
	}
	// Try to delete the session from the remote gateway.
	var reply types.SocketKickReply
	rerr = client.StaticClient(agent.InnerGw).
		Pull(kickUri, types.SocketKickArgs{SessionId: sessionId}, &reply).
		Rerror()
	if reply.Existed {
		return true, nil
	}
	return false, rerr
}

// GetAgent returns agent information.
func GetAgent(sessionId string) (*Agent, *tp.Rerror) {
	key := globalHandler.module.Key(sessionId)
	data, err := globalHandler.redisWithLargeMemory.Get(key).Bytes()
	switch {
	case err == nil:
		a := new(Agent)
		err = json.Unmarshal(data, a)
		if err != nil {
			return nil, newServerRerror(err.Error())
		}
		return a, nil

	case redis.IsRedisNil(err):
		a := new(Agent)
		a.SessionId = sessionId
		a.IsOffline = true
		return a, nil

	default:
		return nil, newServerRerror(err.Error())
	}
}

var nilAgents = &Agents{Agents: []*Agent{}}

// QueryAgent queries agent information in batches.
func QueryAgent(sessionIds []string) (*Agents, *tp.Rerror) {
	if len(sessionIds) == 0 {
		return nilAgents, nil
	}
	var keys = make([]string, len(sessionIds))
	for i, sessionId := range sessionIds {
		keys[i] = globalHandler.module.Key(sessionId)
	}
	rets, err := globalHandler.redisWithLargeMemory.MGet(keys...).Result()
	if err != nil {
		return nil, newServerRerror(err.Error())
	}
	agents := make([]*Agent, len(rets))
	for i, r := range rets {
		a := new(Agent)
		if s, ok := r.(string); ok {
			json.Unmarshal(goutil.StringToBytes(s), a)
		} else {
			a.SessionId = sessionIds[i]
			a.IsOffline = true
		}
		agents[i] = a
	}
	return &Agents{Agents: agents}, nil
}

const (
	// EVENT_ONLINE agent online event
	EVENT_ONLINE = "ONLINE"
	// EVENT_OFFLINE agent offline event
	EVENT_OFFLINE = "OFFLINE"
)

var (
	subscribeInit    sync.Once
	subscribeChannel chan *AgentNews
)

// AgentNews agent online/offline message
type AgentNews struct {
	SessionId string
	Event     string
}

func createAgentNews(sessionId string, event string) string {
	return sessionId + "," + event
}

func parseAgentNews(msg string) (*AgentNews, error) {
	a := strings.Split(msg, ",")
	if len(a) != 2 {
		return nil, fmt.Errorf("The format of the agent news is wrong: %s", msg)
	}
	return &AgentNews{
		SessionId: a[0],
		Event:     a[1],
	}, nil
}

// IsOnline returns whether it is an online message.
func (a *AgentNews) IsOnline() bool {
	return a.Event == EVENT_ONLINE
}

// IsOffline returns whether it is an offline message.
func (a *AgentNews) IsOffline() bool {
	return a.Event == EVENT_OFFLINE
}

// Subscribe subscribes agent news from redis.
func Subscribe() <-chan *AgentNews {
	subscribeInit.Do(func() {
		pubSub := globalHandler.redisWithPublishCmd.Subscribe(agentChannel)
		subscribeChannel = make(chan *AgentNews, 100)
		go func() {
			for msg := range pubSub.Channel() {
				news, err := parseAgentNews(msg.Payload)
				if err != nil {
					tp.Errorf("%s", err.Error())
					continue
				}
				subscribeChannel <- news
			}
		}()
	})
	return subscribeChannel
}
