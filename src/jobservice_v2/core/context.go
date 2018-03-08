package core

import (
	"context"
	"sync"
)

//BaseContext keep some sharable materials.
//The system context.Context interface is also included.
type BaseContext struct {
	//The system context with cancel capability.
	SystemContext context.Context

	//Coordination signal
	WG *sync.WaitGroup
}
