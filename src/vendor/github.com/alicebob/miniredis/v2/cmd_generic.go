// Commands from https://redis.io/commands#generic

package miniredis

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/alicebob/miniredis/v2/server"
)

// commandsGeneric handles EXPIRE, TTL, PERSIST, &c.
func commandsGeneric(m *Miniredis) {
	m.srv.Register("COPY", m.cmdCopy)
	m.srv.Register("DEL", m.cmdDel)
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
	m.srv.Register("TOUCH", m.cmdTouch)
	m.srv.Register("TTL", m.cmdTTL)
	m.srv.Register("TYPE", m.cmdType)
	m.srv.Register("SCAN", m.cmdScan)
	// SORT
	m.srv.Register("UNLINK", m.cmdDel)
}

// generic expire command for EXPIRE, PEXPIRE, EXPIREAT, PEXPIREAT
// d is the time unit. If unix is set it'll be seen as a unixtimestamp and
// converted to a duration.
func makeCmdExpire(m *Miniredis, unix bool, d time.Duration) func(*server.Peer, string, []string) {
	return func(c *server.Peer, cmd string, args []string) {
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

		var opts struct {
			key   string
			value int
			nx    bool
			xx    bool
			gt    bool
			lt    bool
		}
		opts.key = args[0]
		if ok := optInt(c, args[1], &opts.value); !ok {
			return
		}
		args = args[2:]
		for len(args) > 0 {
			switch strings.ToLower(args[0]) {
			case "nx":
				opts.nx = true
			case "xx":
				opts.xx = true
			case "gt":
				opts.gt = true
			case "lt":
				opts.lt = true
			default:
				setDirty(c)
				c.WriteError(fmt.Sprintf("ERR Unsupported option %s", args[0]))
				return
			}
			args = args[1:]
		}
		if opts.gt && opts.lt {
			setDirty(c)
			c.WriteError("ERR GT and LT options at the same time are not compatible")
			return
		}
		if opts.nx && (opts.xx || opts.gt || opts.lt) {
			setDirty(c)
			c.WriteError("ERR NX and XX, GT or LT options at the same time are not compatible")
			return
		}

		withTx(m, c, func(c *server.Peer, ctx *connCtx) {
			db := m.db(ctx.selectedDB)

			// Key must be present.
			if _, ok := db.keys[opts.key]; !ok {
				c.WriteInt(0)
				return
			}

			oldTTL, ok := db.ttl[opts.key]

			var newTTL time.Duration
			if unix {
				newTTL = m.at(opts.value, d)
			} else {
				newTTL = time.Duration(opts.value) * d
			}

			// > NX -- Set expiry only when the key has no expiry
			if opts.nx && ok {
				c.WriteInt(0)
				return
			}
			// > XX -- Set expiry only when the key has an existing expiry
			if opts.xx && !ok {
				c.WriteInt(0)
				return
			}
			// > GT -- Set expiry only when the new expiry is greater than current one
			// (no exp == infinity)
			if opts.gt && (!ok || newTTL <= oldTTL) {
				c.WriteInt(0)
				return
			}
			// > LT -- Set expiry only when the new expiry is less than current one
			if opts.lt && ok && newTTL > oldTTL {
				c.WriteInt(0)
				return
			}
			db.ttl[opts.key] = newTTL
			db.keyVersion[opts.key]++
			db.checkTTL(opts.key)
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

	var opts struct {
		key      string
		targetDB int
	}

	opts.key = args[0]
	opts.targetDB, _ = strconv.Atoi(args[1])

	withTx(m, c, func(c *server.Peer, ctx *connCtx) {
		if ctx.selectedDB == opts.targetDB {
			c.WriteError("ERR source and destination objects are the same")
			return
		}
		db := m.db(ctx.selectedDB)
		targetDB := m.db(opts.targetDB)

		if !db.move(opts.key, targetDB) {
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

	opts := struct {
		from string
		to   string
	}{
		from: args[0],
		to:   args[1],
	}

	withTx(m, c, func(c *server.Peer, ctx *connCtx) {
		db := m.db(ctx.selectedDB)

		if !db.exists(opts.from) {
			c.WriteError(msgKeyNotFound)
			return
		}

		db.rename(opts.from, opts.to)
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

	opts := struct {
		from string
		to   string
	}{
		from: args[0],
		to:   args[1],
	}

	withTx(m, c, func(c *server.Peer, ctx *connCtx) {
		db := m.db(ctx.selectedDB)

		if !db.exists(opts.from) {
			c.WriteError(msgKeyNotFound)
			return
		}

		if db.exists(opts.to) {
			c.WriteInt(0)
			return
		}

		db.rename(opts.from, opts.to)
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

	var opts struct {
		cursor    int
		withMatch bool
		match     string
		withType  bool
		_type     string
	}

	if ok := optIntErr(c, args[0], &opts.cursor, msgInvalidCursor); !ok {
		return
	}
	args = args[1:]

	// MATCH, COUNT and TYPE options
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
			opts.withMatch = true
			opts.match, args = args[1], args[2:]
			continue
		}
		if strings.ToLower(args[0]) == "type" {
			if len(args) < 2 {
				setDirty(c)
				c.WriteError(msgSyntaxError)
				return
			}
			opts.withType = true
			opts._type, args = strings.ToLower(args[1]), args[2:]
			continue
		}
		setDirty(c)
		c.WriteError(msgSyntaxError)
		return
	}

	withTx(m, c, func(c *server.Peer, ctx *connCtx) {
		db := m.db(ctx.selectedDB)
		// We return _all_ (matched) keys every time.

		if opts.cursor != 0 {
			// Invalid cursor.
			c.WriteLen(2)
			c.WriteBulk("0") // no next cursor
			c.WriteLen(0)    // no elements
			return
		}

		var keys []string

		if opts.withType {
			keys = make([]string, 0)
			for k, t := range db.keys {
				// type must be given exactly; no pattern matching is performed
				if t == opts._type {
					keys = append(keys, k)
				}
			}
			sort.Strings(keys) // To make things deterministic.
		} else {
			keys = db.allKeys()
		}

		if opts.withMatch {
			keys, _ = matchKeys(keys, opts.match)
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
