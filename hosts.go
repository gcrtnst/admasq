package main

import (
	"bufio"
	"bytes"
	"io"
	"net/netip"
)

type HostsParser struct {
	Line  int
	IP    netip.Addr
	Hosts []string
	Err   error

	s    *bufio.Scanner
	lnum int
}

func NewHostsParser(r io.Reader) *HostsParser {
	s := bufio.NewScanner(r)
	s.Split(bufio.ScanLines)
	return &HostsParser{s: s}
}

func (p *HostsParser) Parse() bool {
	for p.s.Scan() {
		p.lnum++

		line := p.s.Bytes()
		ip, hs, err := ParseHostsLine(line)
		if err != nil {
			p.Line = p.lnum
			p.IP = netip.Addr{}
			p.Hosts = nil
			p.Err = &ResourceError{Line: p.Line, Err: err}
			return true
		}
		if ip.IsValid() || len(hs) > 0 {
			p.Line = p.lnum
			p.IP = ip
			p.Hosts = hs
			p.Err = nil
			return true
		}
	}

	p.Line = p.lnum
	p.IP = netip.Addr{}
	p.Hosts = nil
	p.Err = p.s.Err()
	return false
}

func ParseHostsLine(line []byte) (netip.Addr, []string, error) {
	var ip netip.Addr
	var hs []string
	buf := line

	fieldIdx := 0
	for {
		fieldLen := bytes.IndexAny(buf, " \t#")
		if fieldLen < 0 {
			fieldLen = len(buf)
		}

		if fieldLen > 0 {
			field := buf[:fieldLen]
			if fieldIdx == 0 {
				var err error
				ip, err = netip.ParseAddr(string(field))
				if err != nil {
					return netip.Addr{}, nil, err
				}
			} else {
				h := string(field)
				hs = append(hs, h)
			}
			fieldIdx++
		}

		if fieldLen >= len(buf) || buf[fieldLen] == '#' {
			break
		}
		buf = buf[fieldLen+1:]
	}
	return ip, hs, nil
}

type HostsIPError struct {
	IP netip.Addr
}

func (e *HostsIPError) Error() string {
	return e.IP.String() + " is neither a loopback address nor an unspecified address"
}
