// Commands from https://redis.io/commands#server

package miniredis

import (
	"strconv"
	"strings"

	"github.com/alicebob/miniredis/v2/server"
)

func commandsServer(m *Miniredis) {
	m.srv.Register("COMMAND", m.cmdCommand)
	m.srv.Register("DBSIZE", m.cmdDbsize)
	m.srv.Register("FLUSHALL", m.cmdFlushall)
	m.srv.Register("FLUSHDB", m.cmdFlushdb)
	m.srv.Register("INFO", m.cmdInfo)
	m.srv.Register("TIME", m.cmdTime)
}

// DBSIZE
func (m *Miniredis) cmdDbsize(c *server.Peer, cmd string, args []string) {
	if len(args) > 0 {
		setDirty(c)
		c.WriteError(errWrongNumber(cmd))
		return
	}
	if !m.handleAuth(c) {
		return
	}
	if m.checkPubsub(c, cmd) {
		return
	}

	withTx(m, c, func(c *server.Peer, ctx *connCtx) {
		db := m.db(ctx.selectedDB)

		c.WriteInt(len(db.keys))
	})
}

// FLUSHALL
func (m *Miniredis) cmdFlushall(c *server.Peer, cmd string, args []string) {
	if len(args) > 0 && strings.ToLower(args[0]) == "async" {
		args = args[1:]
	}
	if len(args) > 0 {
		setDirty(c)
		c.WriteError(msgSyntaxError)
		return
	}
	if !m.handleAuth(c) {
		return
	}
	if m.checkPubsub(c, cmd) {
		return
	}

	withTx(m, c, func(c *server.Peer, ctx *connCtx) {
		m.flushAll()
		c.WriteOK()
	})
}

// FLUSHDB
func (m *Miniredis) cmdFlushdb(c *server.Peer, cmd string, args []string) {
	if len(args) > 0 && strings.ToLower(args[0]) == "async" {
		args = args[1:]
	}
	if len(args) > 0 {
		setDirty(c)
		c.WriteError(msgSyntaxError)
		return
	}
	if !m.handleAuth(c) {
		return
	}
	if m.checkPubsub(c, cmd) {
		return
	}

	withTx(m, c, func(c *server.Peer, ctx *connCtx) {
		m.db(ctx.selectedDB).flush()
		c.WriteOK()
	})
}

// TIME
func (m *Miniredis) cmdTime(c *server.Peer, cmd string, args []string) {
	if len(args) > 0 {
		setDirty(c)
		c.WriteError(errWrongNumber(cmd))
		return
	}
	if !m.handleAuth(c) {
		return
	}
	if m.checkPubsub(c, cmd) {
		return
	}

	withTx(m, c, func(c *server.Peer, ctx *connCtx) {
		now := m.effectiveNow()
		nanos := now.UnixNano()
		seconds := nanos / 1_000_000_000
		microseconds := (nanos / 1_000) % 1_000_000

		c.WriteLen(2)
		c.WriteBulk(strconv.FormatInt(seconds, 10))
		c.WriteBulk(strconv.FormatInt(microseconds, 10))
	})
}
