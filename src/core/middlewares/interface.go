package middlewares

import "github.com/justinas/alice"

type ChainCreator interface {
	Create(middlewares []string) *alice.Chain
}
