package mac2org

import (
	"bufio"
	"context"
	"fmt"
	"iter"
	"log"
	"maps"
	"net"
	"os"

	gm "github.com/google/gopacket/macs"
)

const UnknownOrg string = "(unknown)"

type MacPrefixToOrg func([3]byte) string

type MacPrefixToOrgMap map[[3]byte]string

func (m MacPrefixToOrgMap) ToMapper(unknown string) MacPrefixToOrg {
	return func(prefix [3]byte) string {
		val, found := m[prefix]
		switch found {
		case true:
			return val
		default:
			return unknown
		}
	}
}

func (m MacPrefixToOrgMap) ToMapperDefault() MacPrefixToOrg {
	return m.ToMapper(UnknownOrg)
}

var macPrefix2orgMapDefault MacPrefixToOrgMap = gm.ValidMACPrefixMap

var MacPrefix2orgMapDefault MacPrefixToOrgMap = macPrefix2orgMapDefault

var MacPrefixToOrgDefault MacPrefixToOrg = MacPrefix2orgMapDefault.
	ToMapperDefault()

func (f MacPrefixToOrg) MacToOrg(mac net.HardwareAddr) string {
	var buf [3]byte
	copy(buf[:], mac)
	return f(buf)
}

func (f MacPrefixToOrg) MacStringToOrg(addr string) string {
	parsed, e := net.ParseMAC(addr)
	switch e {
	case nil:
		return f.MacToOrg(parsed)
	default:
		return f.MacToOrg(nil)
	}
}

func (f MacPrefixToOrg) MacStringsToOrgStrings(
	macs iter.Seq[string],
) iter.Seq[string] {
	return func(yield func(string) bool) {
		for addr := range macs {
			var org string = f.MacStringToOrg(addr)
			if !yield(org) {
				return
			}
		}
	}
}

func (f MacPrefixToOrg) StdinToMacStringsToStdout(ctx context.Context) error {
	var i iter.Seq[string] = func(
		yield func(string) bool,
	) {
		var s *bufio.Scanner = bufio.NewScanner(os.Stdin)
		for s.Scan() {
			var line string = s.Text()
			if !yield(line) {
				return
			}

			e := s.Err()
			if nil != e {
				log.Printf("error while scanning mac addr strings: %v\n", e)
			}
		}
	}

	var bw *bufio.Writer = bufio.NewWriter(os.Stdout)

	for addr := range i {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		var org string = f.MacStringToOrg(addr)

		_, e := fmt.Fprintf(bw, "%s: %s\n", addr, org)
		if nil != e {
			return e
		}
	}

	return bw.Flush()
}

func (m MacPrefixToOrgMap) GetNames() iter.Seq[string] {
	return maps.Values(m)
}

var GetNames func() iter.Seq[string] = MacPrefix2orgMapDefault.GetNames
