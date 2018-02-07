package tcp

import (
	tp "github.com/henrylee2cn/teleport"
	"github.com/henrylee2cn/teleport/plugin"
	"github.com/xiaoenai/ants/gateway/auth"
)

type connTab struct{}

var (
	connTabPlugin                         = new(connTab)
	_             tp.PostDisconnectPlugin = new(connTab)
)

func (c *connTab) Name() string {
	return "connTab"
}

func (c *connTab) PostDisconnect(sess tp.BaseSession) *tp.Rerror {
	return c.logoff(sess)
}

func (c *connTab) logon(accessToken string, sess plugin.AuthSession) *tp.Rerror {
	tp.Debugf("verify-auth: id: %s, info: %s", sess.Id(), accessToken)
	token, rerr := auth.Verify(accessToken)
	if rerr != nil {
		return rerr
	}
	// manage session
	// TODO

	if len(token.Id) > 0 {
		sess.SetId(token.Id)
	}
	return nil
}

func (c *connTab) logoff(sess tp.BaseSession) *tp.Rerror {
	// manage session
	// TODO
	tp.Tracef("[-CONN] ip: %s, id: %s", sess.RemoteIp(), sess.Id())
	return nil
}
