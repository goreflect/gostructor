// Command gostructor-dump prints the resolution report served by a service that
// enabled gostructor's DumpTCP debug endpoint (WithDebugDump(gostructor.DumpTCP)
// or GOSTRUCTOR_DEBUG_DUMP). It connects, copies the report to stdout, and
// exits — the same thing `nc host port` does, with a sane default address and
// timeout, and convenient to run over `kubectl exec`.
//
// Usage:
//
//	gostructor-dump [addr]
//
// addr defaults to 127.0.0.1:6555, the loopback address DumpTCP binds by
// default. A bare "host:port" or ":port" is accepted.
package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"time"
)

const defaultAddr = "127.0.0.1:6555"

func main() {
	addr := defaultAddr
	switch len(os.Args) {
	case 1:
	case 2:
		if os.Args[1] == "-h" || os.Args[1] == "--help" {
			fmt.Fprintf(os.Stderr, "usage: %s [addr]\n\naddr defaults to %s\n", os.Args[0], defaultAddr)
			return
		}
		addr = os.Args[1]
	default:
		fmt.Fprintf(os.Stderr, "usage: %s [addr]\n", os.Args[0])
		os.Exit(2)
	}

	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		fmt.Fprintf(os.Stderr, "gostructor-dump: cannot connect to %s: %v\n", addr, err)
		os.Exit(1)
	}
	defer conn.Close()

	_ = conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	if _, err := io.Copy(os.Stdout, conn); err != nil {
		fmt.Fprintf(os.Stderr, "gostructor-dump: read from %s failed: %v\n", addr, err)
		os.Exit(1)
	}
}
