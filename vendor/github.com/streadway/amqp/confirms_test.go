package amqp

import (
	"testing"
	"time"
)

func TestConfirmOneResequences(t *testing.T) {
	var (
		fixtures = []Confirmation{
			{1, true},
			{2, false},
			{3, true},
		}
		c = newConfirms()
		l = make(chan Confirmation, len(fixtures))
	)

	c.Listen(l)

	for i, _ := range fixtures {
		if want, got := uint64(i+1), c.Publish(); want != got {
			t.Fatalf("expected publish to return the 1 based delivery tag published, want: %d, got: %d", want, got)
		}
	}

	c.One(fixtures[1])
	c.One(fixtures[2])

	select {
	case confirm := <-l:
		t.Fatalf("expected to wait in order to properly resequence results, got: %+v", confirm)
	default:
	}

	c.One(fixtures[0])

	for i, fix := range fixtures {
		if want, got := fix, <-l; want != got {
			t.Fatalf("expected to return confirmations in sequence for %d, want: %+v, got: %+v", i, want, got)
		}
	}
}

func TestConfirmMultipleResequences(t *testing.T) {
	var (
		fixtures = []Confirmation{
			{1, true},
			{2, true},
			{3, true},
			{4, true},
		}
		c = newConfirms()
		l = make(chan Confirmation, len(fixtures))
	)
	c.Listen(l)

	for _, _ = range fixtures {
		c.Publish()
	}

	c.Multiple(fixtures[len(fixtures)-1])

	for i, fix := range fixtures {
		if want, got := fix, <-l; want != got {
			t.Fatalf("expected to confirm multiple in sequence for %d, want: %+v, got: %+v", i, want, got)
		}
	}
}

func BenchmarkSequentialBufferedConfirms(t *testing.B) {
	var (
		c = newConfirms()
		l = make(chan Confirmation, 10)
	)

	c.Listen(l)

	for i := 0; i < t.N; i++ {
		if i > cap(l)-1 {
			<-l
		}
		c.One(Confirmation{c.Publish(), true})
	}
}

func TestConfirmsIsThreadSafe(t *testing.T) {
	const count = 1000
	const timeout = 5 * time.Second
	var (
		c    = newConfirms()
		l    = make(chan Confirmation)
		pub  = make(chan Confirmation)
		done = make(chan Confirmation)
		late = time.After(timeout)
	)

	c.Listen(l)

	for i := 0; i < count; i++ {
		go func() { pub <- Confirmation{c.Publish(), true} }()
	}

	for i := 0; i < count; i++ {
		go func() { c.One(<-pub) }()
	}

	for i := 0; i < count; i++ {
		go func() { done <- <-l }()
	}

	for i := 0; i < count; i++ {
		select {
		case <-done:
		case <-late:
			t.Fatalf("expected all publish/confirms to finish after %s", timeout)
		}
	}
}
