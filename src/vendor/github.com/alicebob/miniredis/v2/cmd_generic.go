// Commands from https://redis.io/commands#generic

package miniredis

import (
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/alicebob/miniredis/v2/server"
)

// commandsGeneric handles EXPIRE, TTL, PERSIST, &c.
func commandsGeneric(m *Miniredis) {
	m.srv.Register("DEL", m.cmdDel)
	m.srv.Register("UNLINK", m.cmdDel)
	// DUMP
	m.srv.Register("EXISTS", m.cmdExists)
	m.srv.Register("EXPIRE", makeCmdExpire(m, false, time.Second))
	m.srv.Register("EXPIREAT", makeCmdExpire(m, true, time.Second))
	m.srv.Register("KEYS", m.cmdKeys)
	// MIGRATE
	m.srv.Register("MOVE", m.cmdMove)
	// OBJECT
	m.srv.Register("PERSIST", m.cmdPersist)
	m.srv.Register("PEXPIRE", makeCmdExpire(m, false, time.Millisecond))
	m.srv.Register("PEXPIREAT", makeCmdExpire(m, true, time.Millisecond))
	m.srv.Register("PTTL", m.cmdPTTL)
	m.srv.Register("RANDOMKEY", m.cmdRandomkey)
	m.srv.Register("RENAME", m.cmdRename)
	m.srv.Register("RENAMENX", m.cmdRenamenx)
	// RESTORE
	// SORT
	m.srv.Register("TOUCH", m.cmdTouch)
	m.srv.Register("TTL", m.cmdTTL)
	m.srv.Register("TYPE", m.cmdType)
	m.srv.Register("SCAN", m.cmdScan)
	m.srv.Register("COPY", m.cmdCopy)
}

// generic expire command for EXPIRE, PEXPIRE, EXPIREAT, PEXPIREAT
// d is the time unit. If unix is set it'll be seen as a unixtimestamp and
// converted to a duration.
func makeCmdExpire(m *Miniredis, unix bool, d time.Duration) func(*server.Peer, string, []string) {
	return func(c *server.Peer, cmd string, args []string) {
		if len(args) != 2 {
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

		key := args[0]
		value := args[1]
		i, err := strconv.Atoi(value)
		if err != nil {
			setDirty(c)
			c.WriteError(msgInvalidInt)
			return
		}

		withTx(m, c, func(c *server.Peer, ctx *connCtx) {
			db := m.db(ctx.selectedDB)

			// Key must be present.
			if _, ok := db.keys[key]; !ok {
				c.WriteInt(0)
				return
			}
			if unix {
				db.ttl[key] = m.at(i, d)
			} else {
				db.ttl[key] = time.Duration(i) * d
			}
			db.keyVersion[key]++
			db.checkTTL(key)
			c.WriteInt(1)
		})
	}
}

// TOUCH
func (m *Miniredis) cmdTouch(c *server.Peer, cmd string, args []string) {
	if !m.handleAuth(c) {
		return
	}
	if m.checkPubsub(c, cmd) {
		return
	}

	if len(args) == 0 {
		setDirty(c)
		c.WriteError(errWrongNumber(cmd))
		return
	}

	withTx(m, c, func(c *server.Peer, ctx *connCtx) {
		db := m.db(ctx.selectedDB)

		count := 0
		for _, key := range args {
			if db.exists(key) {
				count++
			}
		}
		c.WriteInt(count)
	})
}

// TTL
func (m *Miniredis) cmdTTL(c *server.Peer, cmd string, args []string) {
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

	key := args[0]

	withTx(m, c, func(c *server.Peer, ctx *connCtx) {
		db := m.db(ctx.selectedDB)

		if _, ok := db.keys[key]; !ok {
			// No such key
			c.WriteInt(-2)
			return
		}

		v, ok := db.ttl[key]
		if !ok {
			// no expire value
			c.WriteInt(-1)
			return
		}
		c.WriteInt(int(v.Seconds()))
	})
}

// PTTL
func (m *Miniredis) cmdPTTL(c *server.Peer, cmd string, args []string) {
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

	key := args[0]

	withTx(m, c, func(c *server.Peer, ctx *connCtx) {
		db := m.db(ctx.selectedDB)

		if _, ok := db.keys[key]; !ok {
			// no such key
			c.WriteInt(-2)
			return
		}

		v, ok := db.ttl[key]
		if !ok {
			// no expire value
			c.WriteInt(-1)
			return
		}
		c.WriteInt(int(v.Nanoseconds() / 1000000))
	})
}

// PERSIST
func (m *Miniredis) cmdPersist(c *server.Peer, cmd string, args []string) {
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

	key := args[0]

	withTx(m, c, func(c *server.Peer, ctx *connCtx) {
		db := m.db(ctx.selectedDB)

		if _, ok := db.keys[key]; !ok {
			// no such key
			c.WriteInt(0)
			return
		}

		if _, ok := db.ttl[key]; !ok {
			// no expire value
			c.WriteInt(0)
			return
		}
		delete(db.ttl, key)
		db.keyVersion[key]++
		c.WriteInt(1)
	})
}

// DEL and UNLINK
func (m *Miniredis) cmdDel(c *server.Peer, cmd string, args []string) {
	if !m.handleAuth(c) {
		return
	}
	if m.checkPubsub(c, cmd) {
		return
	}

	if len(args) == 0 {
		setDirty(c)
		c.WriteError(errWrongNumber(cmd))
		return
	}

	withTx(m, c, func(c *server.Peer, ctx *connCtx) {
		db := m.db(ctx.selectedDB)

		count := 0
		for _, key := range args {
			if db.exists(key) {
				count++
			}
			db.del(key, true) // delete expire
		}
		c.WriteInt(count)
	})
}

// TYPE
func (m *Miniredis) cmdType(c *server.Peer, cmd string, args []string) {
	if len(args) != 1 {
		setDirty(c)
		c.WriteError("usage error")
		return
	}
	if !m.handleAuth(c) {
		return
	}
	if m.checkPubsub(c, cmd) {
		return
	}

	key := args[0]

	withTx(m, c, func(c *server.Peer, ctx *connCtx) {
		db := m.db(ctx.selectedDB)

		t, ok := db.keys[key]
		if !ok {
			c.WriteInline("none")
			return
		}

		c.WriteInline(t)
	})
}

// EXISTS
func (m *Miniredis) cmdExists(c *server.Peer, cmd string, args []string) {
	if len(args) < 1 {
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

		found := 0
		for _, k := range args {
			if db.exists(k) {
				found++
			}
		}
		c.WriteInt(found)
	})
}

// MOVE
func (m *Miniredis) cmdMove(c *server.Peer, cmd string, args []string) {
	if len(args) != 2 {
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

	key := args[0]
	targetDB, err := strconv.Atoi(args[1])
	if err != nil {
		targetDB = 0
	}

	withTx(m, c, func(c *server.Peer, ctx *connCtx) {
		if ctx.selectedDB == targetDB {
			c.WriteError("ERR source and destination objects are the same")
			return
		}
		db := m.db(ctx.selectedDB)
		targetDB := m.db(targetDB)

		if !db.move(key, targetDB) {
			c.WriteInt(0)
			return
		}
		c.WriteInt(1)
	})
}

// KEYS
func (m *Miniredis) cmdKeys(c *server.Peer, cmd string, args []string) {
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

	key := args[0]

	withTx(m, c, func(c *server.Peer, ctx *connCtx) {
		db := m.db(ctx.selectedDB)

		keys, _ := matchKeys(db.allKeys(), key)
		c.WriteLen(len(keys))
		for _, s := range keys {
			c.WriteBulk(s)
		}
	})
}

// RANDOMKEY
func (m *Miniredis) cmdRandomkey(c *server.Peer, cmd string, args []string) {
	if len(args) != 0 {
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

		if len(db.keys) == 0 {
			c.WriteNull()
			return
		}
		nr := m.randIntn(len(db.keys))
		for k := range db.keys {
			if nr == 0 {
				c.WriteBulk(k)
				return
			}
			nr--
		}
	})
}

// RENAME
func (m *Miniredis) cmdRename(c *server.Peer, cmd string, args []string) {
	if len(args) != 2 {
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

	from, to := args[0], args[1]

	withTx(m, c, func(c *server.Peer, ctx *connCtx) {
		db := m.db(ctx.selectedDB)

		if !db.exists(from) {
			c.WriteError(msgKeyNotFound)
			return
		}

		db.rename(from, to)
		c.WriteOK()
	})
}

// RENAMENX
func (m *Miniredis) cmdRenamenx(c *server.Peer, cmd string, args []string) {
	if len(args) != 2 {
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

	from, to := args[0], args[1]

	withTx(m, c, func(c *server.Peer, ctx *connCtx) {
		db := m.db(ctx.selectedDB)

		if !db.exists(from) {
			c.WriteError(msgKeyNotFound)
			return
		}

		if db.exists(to) {
			c.WriteInt(0)
			return
		}

		db.rename(from, to)
		c.WriteInt(1)
	})
}

// SCAN
func (m *Miniredis) cmdScan(c *server.Peer, cmd string, args []string) {
	if len(args) < 1 {
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

	cursor, err := strconv.Atoi(args[0])
	if err != nil {
		setDirty(c)
		c.WriteError(msgInvalidCursor)
		return
	}
	args = args[1:]

	// MATCH, COUNT and TYPE options
	var (
		withMatch bool
		match     string
		withType  bool
		_type     string
	)

	for len(args) > 0 {
		if strings.ToLower(args[0]) == "count" {
			// we do nothing with count
			if len(args) < 2 {
				setDirty(c)
				c.WriteError(msgSyntaxError)
				return
			}
			if _, err := strconv.Atoi(args[1]); err != nil {
				setDirty(c)
				c.WriteError(msgInvalidInt)
				return
			}
			args = args[2:]
			continue
		}
		if strings.ToLower(args[0]) == "match" {
			if len(args) < 2 {
				setDirty(c)
				c.WriteError(msgSyntaxError)
				return
			}
			withMatch = true
			match, args = args[1], args[2:]
			continue
		}
		if strings.ToLower(args[0]) == "type" {
			if len(args) < 2 {
				setDirty(c)
				c.WriteError(msgSyntaxError)
				return
			}
			withType = true
			_type, args = strings.ToLower(args[1]), args[2:]
			continue
		}
		setDirty(c)
		c.WriteError(msgSyntaxError)
		return
	}

	withTx(m, c, func(c *server.Peer, ctx *connCtx) {
		db := m.db(ctx.selectedDB)
		// We return _all_ (matched) keys every time.

		if cursor != 0 {
			// Invalid cursor.
			c.WriteLen(2)
			c.WriteBulk("0") // no next cursor
			c.WriteLen(0)    // no elements
			return
		}

		var keys []string

		if withType {
			keys = make([]string, 0)
			for k, t := range db.keys {
				// type must be given exactly; no pattern matching is performed
				if t == _type {
					keys = append(keys, k)
				}
			}
			sort.Strings(keys) // To make things deterministic.
		} else {
			keys = db.allKeys()
		}

		if withMatch {
			keys, _ = matchKeys(keys, match)
		}

		c.WriteLen(2)
		c.WriteBulk("0") // no next cursor
		c.WriteLen(len(keys))
		for _, k := range keys {
			c.WriteBulk(k)
		}
	})
}

// COPY
func (m *Miniredis) cmdCopy(c *server.Peer, cmd string, args []string) {
	if len(args) < 2 {
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

	var opts = struct {
		from          string
		to            string
		destinationDB int
		replace       bool
	}{
		destinationDB: -1,
	}

	opts.from, opts.to, args = args[0], args[1], args[2:]
	for len(args) > 0 {
		switch strings.ToLower(args[0]) {
		case "db":
			if len(args) < 2 {
				setDirty(c)
				c.WriteError(msgSyntaxError)
				return
			}
			db, err := strconv.Atoi(args[1])
			if err != nil {
				setDirty(c)
				c.WriteError(msgInvalidInt)
				return
			}
			if db < 0 {
				setDirty(c)
				c.WriteError(msgDBIndexOutOfRange)
				return
			}
			opts.destinationDB = db
			args = args[2:]
		case "replace":
			opts.replace = true
			args = args[1:]
		default:
			setDirty(c)
			c.WriteError(msgSyntaxError)
			return
		}
	}

	withTx(m, c, func(c *server.Peer, ctx *connCtx) {
		fromDB, toDB := ctx.selectedDB, opts.destinationDB
		if toDB == -1 {
			toDB = fromDB
		}

		if fromDB == toDB && opts.from == opts.to {
			c.WriteError("ERR source and destination objects are the same")
			return
		}

		if !m.db(fromDB).exists(opts.from) {
			c.WriteInt(0)
			return
		}

		if !opts.replace {
			if m.db(toDB).exists(opts.to) {
				c.WriteInt(0)
				return
			}
		}

		m.copy(m.db(fromDB), opts.from, m.db(toDB), opts.to)
		c.WriteInt(1)
	})
}
