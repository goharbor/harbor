# Redsync

[![Build Status](https://drone.io/github.com/go-redsync/redsync/status.png)](https://drone.io/github.com/go-redsync/redsync/latest)

Redsync provides a Redis-based distributed mutual exclusion lock implementation for Go as described in [this post](http://redis.io/topics/distlock). A reference library (by [antirez](https://github.com/antirez)) for Ruby is available at [github.com/antirez/redlock-rb](https://github.com/antirez/redlock-rb).

## Installation

Install Redsync using the go get command:

    $ go get gopkg.in/redsync.v1

The only dependencies are the Go distribution and [Redigo](github.com/garyburd/redigo).

## Documentation

- [Reference](https://godoc.org/gopkg.in/redsync.v1)

## Contributing

Contributions are welcome.

## License

Redsync is available under the [BSD (3-Clause) License](https://opensource.org/licenses/BSD-3-Clause).

## Disclaimer

This code implements an algorithm which is currently a proposal, it was not formally analyzed. Make sure to understand how it works before using it in production environments.
