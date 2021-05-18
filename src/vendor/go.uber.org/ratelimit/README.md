# Go rate limiter [![GoDoc][doc-img]][doc] [![Coverage Status][cov-img]][cov] ![test][test-img]

This package provides a Golang implementation of the leaky-bucket rate limit algorithm.
This implementation refills the bucket based on the time elapsed between
requests instead of requiring an interval clock to fill the bucket discretely.

Create a rate limiter with a maximum number of operations to perform per second.
Call Take() before each operation. Take will sleep until you can continue.

```go
import (
	"fmt"
	"time"

	"go.uber.org/ratelimit"
)

func main() {
    rl := ratelimit.New(100) // per second

    prev := time.Now()
    for i := 0; i < 10; i++ {
        now := rl.Take()
        fmt.Println(i, now.Sub(prev))
        prev = now
    }

    // Output:
    // 0 0
    // 1 10ms
    // 2 10ms
    // 3 10ms
    // 4 10ms
    // 5 10ms
    // 6 10ms
    // 7 10ms
    // 8 10ms
    // 9 10ms
}
```

[cov-img]: https://codecov.io/gh/uber-go/ratelimit/branch/master/graph/badge.svg?token=zhLeUjjrm2
[cov]: https://codecov.io/gh/uber-go/ratelimit
[doc-img]: https://pkg.go.dev/badge/go.uber.org/ratelimit
[doc]: https://pkg.go.dev/go.uber.org/ratelimit
[test-img]: https://github.com/uber-go/ratelimit/workflows/test/badge.svg
