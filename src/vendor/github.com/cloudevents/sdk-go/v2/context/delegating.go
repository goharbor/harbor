package context

import "context"

type valuesDelegating struct {
	context.Context
	parent context.Context
}

// ValuesDelegating wraps a child and parent context. It will perform Value()
// lookups first on the child, and then fall back to the child. All other calls
// go solely to the child context.
func ValuesDelegating(child, parent context.Context) context.Context {
	return &valuesDelegating{
		Context: child,
		parent:  parent,
	}
}

func (c *valuesDelegating) Value(key interface{}) interface{} {
	if val := c.Context.Value(key); val != nil {
		return val
	}
	return c.parent.Value(key)
}
