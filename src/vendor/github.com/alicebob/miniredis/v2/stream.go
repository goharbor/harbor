// Basic stream implementation.

package miniredis

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"
)

// a Stream is a list of entries, lowest ID (oldest) first, and all "groups".
type streamKey struct {
	entries         []StreamEntry
	groups          map[string]*streamGroup
	lastAllocatedID string
}

// a StreamEntry is an entry in a stream. The ID is always of the form
// "123-123".
// Values is an ordered list of key-value pairs.
type StreamEntry struct {
	ID     string
	Values []string
}

type streamGroup struct {
	stream    *streamKey
	lastID    string
	pending   []pendingEntry
	consumers map[string]consumer
}

type consumer struct {
	// TODO: "last seen" timestamp
}

type pendingEntry struct {
	id            string
	consumer      string
	deliveryCount int
	lastDelivery  time.Time
}

func newStreamKey() *streamKey {
	return &streamKey{
		groups: map[string]*streamGroup{},
	}
}

func (s *streamKey) generateID(now time.Time) string {
	ts := uint64(now.UnixNano()) / 1_000_000

	next := fmt.Sprintf("%d-%d", ts, 0)
	if s.lastAllocatedID != "" && streamCmp(s.lastAllocatedID, next) >= 0 {
		last, _ := parseStreamID(s.lastAllocatedID)
		next = fmt.Sprintf("%d-%d", last[0], last[1]+1)
	}

	lastID := s.lastID()
	if streamCmp(lastID, next) >= 0 {
		last, _ := parseStreamID(lastID)
		next = fmt.Sprintf("%d-%d", last[0], last[1]+1)
	}

	s.lastAllocatedID = next
	return next
}

func (s *streamKey) lastID() string {
	if len(s.entries) == 0 {
		return "0-0"
	}

	return s.entries[len(s.entries)-1].ID
}

func (s *streamKey) copy() *streamKey {
	cpy := &streamKey{
		entries: s.entries,
	}
	groups := map[string]*streamGroup{}
	for k, v := range s.groups {
		gr := v.copy()
		gr.stream = cpy
		groups[k] = gr
	}
	cpy.groups = groups
	return cpy
}

func parseStreamID(id string) ([2]uint64, error) {
	var (
		res [2]uint64
		err error
	)
	parts := strings.SplitN(id, "-", 2)
	res[0], err = strconv.ParseUint(parts[0], 10, 64)
	if err != nil {
		return res, errors.New(msgInvalidStreamID)
	}
	if len(parts) == 2 {
		res[1], err = strconv.ParseUint(parts[1], 10, 64)
		if err != nil {
			return res, errors.New(msgInvalidStreamID)
		}
	}
	return res, nil
}

// compares two stream IDs (of the full format: "123-123"). Returns: -1, 0, 1
// The given IDs should be valid stream IDs.
func streamCmp(a, b string) int {
	ap, _ := parseStreamID(a)
	bp, _ := parseStreamID(b)

	switch {
	case ap[0] < bp[0]:
		return -1
	case ap[0] > bp[0]:
		return 1
	case ap[1] < bp[1]:
		return -1
	case ap[1] > bp[1]:
		return 1
	default:
		return 0
	}
}

// formatStreamID makes a full id ("42-42") out of a partial one ("42")
func formatStreamID(id string) (string, error) {
	var ts [2]uint64
	parts := strings.SplitN(id, "-", 2)

	if len(parts) > 0 {
		p, err := strconv.ParseUint(parts[0], 10, 64)
		if err != nil {
			return "", errInvalidEntryID
		}
		ts[0] = p
	}
	if len(parts) > 1 {
		p, err := strconv.ParseUint(parts[1], 10, 64)
		if err != nil {
			return "", errInvalidEntryID
		}
		ts[1] = p
	}
	return fmt.Sprintf("%d-%d", ts[0], ts[1]), nil
}

func formatStreamRangeBound(id string, start bool, reverse bool) (string, error) {
	if id == "-" {
		return "0-0", nil
	}

	if id == "+" {
		return fmt.Sprintf("%d-%d", uint64(math.MaxUint64), uint64(math.MaxUint64)), nil
	}

	if id == "0" {
		return "0-0", nil
	}

	parts := strings.Split(id, "-")
	if len(parts) == 2 {
		return formatStreamID(id)
	}

	// Incomplete IDs case
	ts, err := strconv.ParseUint(parts[0], 10, 64)
	if err != nil {
		return "", errInvalidEntryID
	}

	if (!start && !reverse) || (start && reverse) {
		return fmt.Sprintf("%d-%d", ts, uint64(math.MaxUint64)), nil
	}

	return fmt.Sprintf("%d-%d", ts, 0), nil
}

func reversedStreamEntries(o []StreamEntry) []StreamEntry {
	newStream := make([]StreamEntry, len(o))
	for i, e := range o {
		newStream[len(o)-i-1] = e
	}
	return newStream
}

func (s *streamKey) createGroup(group, id string) error {
	if _, ok := s.groups[group]; ok {
		return errors.New("BUSYGROUP Consumer Group name already exists")
	}

	if id == "$" {
		id = s.lastID()
	}
	s.groups[group] = &streamGroup{
		stream:    s,
		lastID:    id,
		consumers: map[string]consumer{},
	}
	return nil
}

// streamAdd adds an entry to a stream. Returns the new entry ID.
// If id is empty or "*" the ID will be generated automatically.
// `values` should have an even length.
func (s *streamKey) add(entryID string, values []string, now time.Time) (string, error) {
	if entryID == "" || entryID == "*" {
		entryID = s.generateID(now)
	}

	entryID, err := formatStreamID(entryID)
	if err != nil {
		return "", err
	}
	if entryID == "0-0" {
		return "", errors.New(msgStreamIDZero)
	}
	if streamCmp(s.lastID(), entryID) != -1 {
		return "", errors.New(msgStreamIDTooSmall)
	}

	s.entries = append(s.entries, StreamEntry{
		ID:     entryID,
		Values: values,
	})
	return entryID, nil
}

func (s *streamKey) trim(n int) {
	if len(s.entries) > n {
		s.entries = s.entries[len(s.entries)-n:]
	}
}

// all entries after "id"
func (s *streamKey) after(id string) []StreamEntry {
	pos := sort.Search(len(s.entries), func(i int) bool {
		return streamCmp(id, s.entries[i].ID) < 0
	})
	return s.entries[pos:]
}

// get a stream entry by ID
// Also returns the position in the entries slice, if found.
func (s *streamKey) get(id string) (int, *StreamEntry) {
	pos := sort.Search(len(s.entries), func(i int) bool {
		return streamCmp(id, s.entries[i].ID) <= 0
	})
	if len(s.entries) <= pos || s.entries[pos].ID != id {
		return 0, nil
	}
	return pos, &s.entries[pos]
}

func (g *streamGroup) readGroup(
	now time.Time,
	consumerID,
	id string,
	count int,
	noack bool,
) []StreamEntry {
	if id == ">" {
		// undelivered messages
		msgs := g.stream.after(g.lastID)
		if len(msgs) == 0 {
			return nil
		}

		if count > 0 && len(msgs) > count {
			msgs = msgs[:count]
		}

		if !noack {
			for _, msg := range msgs {
				g.pending = append(g.pending, pendingEntry{
					id:            msg.ID,
					consumer:      consumerID,
					deliveryCount: 1,
					lastDelivery:  now,
				})
			}
		}
		g.consumers[consumerID] = consumer{}
		g.lastID = msgs[len(msgs)-1].ID
		return msgs
	}

	// re-deliver messages from the pending list.
	// con := gr.consumers[consumerID]
	msgs := g.pendingAfter(id)
	var res []StreamEntry
	for i, p := range msgs {
		if p.consumer != consumerID {
			continue
		}
		_, entry := g.stream.get(p.id)
		// not found. Weird?
		if entry == nil {
			continue
		}
		p.deliveryCount += 1
		p.lastDelivery = now
		msgs[i] = p
		res = append(res, *entry)
	}
	return res
}

func (g *streamGroup) ack(ids []string) (int, error) {
	count := 0
	for _, id := range ids {
		if _, err := parseStreamID(id); err != nil {
			return 0, errors.New(msgInvalidStreamID)
		}

		pos := sort.Search(len(g.pending), func(i int) bool {
			return streamCmp(id, g.pending[i].id) <= 0
		})
		if len(g.pending) <= pos || g.pending[pos].id != id {
			continue
		}

		g.pending = append(g.pending[:pos], g.pending[pos+1:]...)
		count++
	}
	return count, nil
}

func (s *streamKey) delete(ids []string) (int, error) {
	count := 0
	for _, id := range ids {
		if _, err := parseStreamID(id); err != nil {
			return 0, errors.New(msgInvalidStreamID)
		}

		i, entry := s.get(id)
		if entry == nil {
			continue
		}

		s.entries = append(s.entries[:i], s.entries[i+1:]...)
		count++
	}
	return count, nil
}

func (g *streamGroup) pendingAfter(id string) []pendingEntry {
	pos := sort.Search(len(g.pending), func(i int) bool {
		return streamCmp(id, g.pending[i].id) < 0
	})
	return g.pending[pos:]
}

func (g *streamGroup) pendingCount(consumer string) int {
	n := 0
	for _, p := range g.pending {
		if p.consumer == consumer {
			n++
		}
	}
	return n
}

func (g *streamGroup) copy() *streamGroup {
	cns := map[string]consumer{}
	for k, v := range g.consumers {
		cns[k] = v
	}
	return &streamGroup{
		// don't copy stream
		lastID:    g.lastID,
		pending:   g.pending,
		consumers: cns,
	}
}
