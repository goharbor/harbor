package work

import (
	"fmt"
	"reflect"
)

// returns an error if the job fails, or there's a panic, or we couldn't reflect correctly.
// if we return an error, it signals we want the job to be retried.
func runJob(job *Job, ctxType reflect.Type, middleware []*middlewareHandler, jt *jobType) (returnCtx reflect.Value, returnError error) {
	returnCtx = reflect.New(ctxType)
	currentMiddleware := 0
	maxMiddleware := len(middleware)

	var next NextMiddlewareFunc
	next = func() error {
		if currentMiddleware < maxMiddleware {
			mw := middleware[currentMiddleware]
			currentMiddleware++
			if mw.IsGeneric {
				return mw.GenericMiddlewareHandler(job, next)
			}
			res := mw.DynamicMiddleware.Call([]reflect.Value{returnCtx, reflect.ValueOf(job), reflect.ValueOf(next)})
			x := res[0].Interface()
			if x == nil {
				return nil
			}
			return x.(error)
		}
		if jt.IsGeneric {
			return jt.GenericHandler(job)
		}
		res := jt.DynamicHandler.Call([]reflect.Value{returnCtx, reflect.ValueOf(job)})
		x := res[0].Interface()
		if x == nil {
			return nil
		}
		return x.(error)
	}

	defer func() {
		if panicErr := recover(); panicErr != nil {
			// err turns out to be interface{}, of actual type "runtime.errorCString"
			// Luckily, the err sprints nicely via fmt.
			errorishError := fmt.Errorf("%v", panicErr)
			logError("runJob.panic", errorishError)
			returnError = errorishError
		}
	}()

	returnError = next()

	return
}
