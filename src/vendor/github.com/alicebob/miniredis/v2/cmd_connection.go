// Commands from https://redis.io/commands#connection

package miniredis

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/alicebob/miniredis/v2/server"
)

func commandsConnection(m *Miniredis) {
	m.srv.Register("AUTH", m.cmdAuth)
	m.srv.Register("ECHO", m.cmdEcho)
	m.srv.Register("HELLO", m.cmdHello)
	m.srv.Register("PING", m.cmdPing)
	m.srv.Register("QUIT", m.cmdQuit)
	m.srv.Register("SELECT", m.cmdSelect)
	m.srv.Register("SWAPDB", m.cmdSwapdb)
}

// PING
func (m *Miniredis) cmdPing(c *server.Peer, cmd string, args []string) {
	if !m.handleAuth(c) {
		return
	}

	if len(args) > 1 {
		setDirty(c)
		c.WriteError(errWrongNumber(cmd))
		return
	}

	payload := ""
	if len(args) > 0 {
		payload = args[0]
	}

	// PING is allowed in subscribed state
	if sub := getCtx(c).subscriber; sub != nil {
		c.Block(func(c *server.Writer) {
			c.WriteLen(2)
			c.WriteBulk("pong")
			c.WriteBulk(payload)
		})
		return
	}

	withTx(m, c, func(c *server.Peer, ctx *connCtx) {
		if payload == "" {
			c.WriteInline("PONG")
			return
		}
		c.WriteBulk(payload)
	})
}

// AUTH
func (m *Miniredis) cmdAuth(c *server.Peer, cmd string, args []string) {
	if len(args) < 1 {
		setDirty(c)
		c.WriteError(errWrongNumber(cmd))
		return
	}

	if len(args) > 2 {
		c.WriteError(msgSyntaxError)
		return
	}
	if m.checkPubsub(c, cmd) {
		return
	}
	if getCtx(c).nested {
		c.WriteError(msgNotFromScripts)
		return
	}
	username := "default"
	pw := args[0]
	if len(args) == 2 {
		username, pw = args[0], args[1]
	}

	withTx(m, c, func(c *server.Peer, ctx *connCtx) {
		if len(m.passwords) == 0 && username == "default" {
			c.WriteError("ERR AUTH <password> called without any password configured for the default user. Are you sure your configuration is correct?")
			return
		}
		setPW, ok := m.passwords[username]
		if !ok {
			c.WriteError("WRONGPASS invalid username-password pair")
			return
		}
		if setPW != pw {
			c.WriteError("WRONGPASS invalid username-password pair")
			return
		}

		ctx.authenticated = true
		c.WriteOK()
	})
}

// HELLO
func (m *Miniredis) cmdHello(c *server.Peer, cmd string, args []string) {
	if len(args) < 1 {
		c.WriteError(errWrongNumber(cmd))
		return
	}

	var opts struct {
		version            int
		username, password string
	}

	versionArg, args := args[0], args[1:]
	var err error
	opts.version, err = strconv.Atoi(versionArg)
	if err != nil {
		c.WriteError("ERR Protocol version is not an integer or out of range")
		return
	}
	switch opts.version {
	case 2, 3:
	default:
		c.WriteError("NOPROTO unsupported protocol version")
		return
	}

	var checkAuth bool
	for len(args) > 0 {
		switch strings.ToUpper(args[0]) {
		case "AUTH":
			if len(args) < 3 {
				c.WriteError(fmt.Sprintf("ERR Syntax error in HELLO option '%s'", args[0]))
				return
			}
			opts.username, opts.password, args = args[1], args[2], args[3:]
			checkAuth = true
		case "SETNAME":
			if len(args) < 2 {
				c.WriteError(fmt.Sprintf("ERR Syntax error in HELLO option '%s'", args[0]))
				return
			}
			_, args = args[1], args[2:]
		default:
			c.WriteError(fmt.Sprintf("ERR Syntax error in HELLO option '%s'", args[0]))
			return
		}
	}

	if len(m.passwords) == 0 && opts.username == "default" {
		// redis ignores legacy "AUTH" if it's not enabled.
		checkAuth = false
	}
	if checkAuth {
		setPW, ok := m.passwords[opts.username]
		if !ok {
			c.WriteError("WRONGPASS invalid username-password pair")
			return
		}
		if setPW != opts.password {
			c.WriteError("WRONGPASS invalid username-password pair")
			return
		}
		getCtx(c).authenticated = true
	}

	c.Resp3 = opts.version == 3

	c.WriteMapLen(7)
	c.WriteBulk("server")
	c.WriteBulk("miniredis")
	c.WriteBulk("version")
	c.WriteBulk("6.0.5")
	c.WriteBulk("proto")
	c.WriteInt(opts.version)
	c.WriteBulk("id")
	c.WriteInt(42)
	c.WriteBulk("mode")
	c.WriteBulk("standalone")
	c.WriteBulk("role")
	c.WriteBulk("master")
	c.WriteBulk("modules")
	c.WriteLen(0)
}

// ECHO
func (m *Miniredis) cmdEcho(c *server.Peer, cmd string, args []string) {
	if len(args) != 1 {
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
		msg := args[0]
		c.WriteBulk(msg)
	})
}

// SELECT
func (m *Miniredis) cmdSelect(c *server.Peer, cmd string, args []string) {
	if len(args) != 1 {
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
		id, err := strconv.Atoi(args[0])
		if err != nil {
			c.WriteError(msgInvalidInt)
			setDirty(c)
			return
		}
		if id < 0 {
			c.WriteError(msgDBIndexOutOfRange)
			setDirty(c)
			return
		}

		ctx.selectedDB = id
		c.WriteOK()
	})
}

// SWAPDB
func (m *Miniredis) cmdSwapdb(c *server.Peer, cmd string, args []string) {
	if len(args) != 2 {
		setDirty(c)
		c.WriteError(errWrongNumber(cmd))
		return
	}
	if !m.handleAuth(c) {
		return
	}

	withTx(m, c, func(c *server.Peer, ctx *connCtx) {
		id1, err := strconv.Atoi(args[0])
		if err != nil {
			c.WriteError("ERR invalid first DB index")
			setDirty(c)
			return
		}
		id2, err := strconv.Atoi(args[1])
		if err != nil {
			c.WriteError("ERR invalid second DB index")
			setDirty(c)
			return
		}
		if id1 < 0 || id2 < 0 {
			c.WriteError(msgDBIndexOutOfRange)
			setDirty(c)
			return
		}

		m.swapDB(id1, id2)

		c.WriteOK()
	})
}

// QUIT
func (m *Miniredis) cmdQuit(c *server.Peer, cmd string, args []string) {
	// QUIT isn't transactionfied and accepts any arguments.
	c.WriteOK()
	c.Close()
}
