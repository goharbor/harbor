// +build darwin dragonfly freebsd linux netbsd openbsd solaris windows

package main

var (
	unix = flags.Bool("unix", false, prettify(`
		Indicates that the server address is the path to a Unix domain socket.`))
)

func init() {
	isUnixSocket = func() bool {
		return *unix
	}
}
