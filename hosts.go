package main

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"net/netip"
)

var ErrMissingHostname = errors.New("missing hostname field")

type HostsLoader struct {
	p *HostsParser

	hs []string
	i  int

	f   Filter
	err error
}

func NewHostsLoader(r io.Reader) *HostsLoader {
	p := NewHostsParser(r)
	return &HostsLoader{p: p}
}

func (l *HostsLoader) Load() bool {
	if l.i+1 < len(l.hs) {
		l.i++
		l.setDomain(l.hs[l.i])
		return true
	}
	l.hs = nil
	l.i = 0

	for l.p.Parse() {
		if l.p.Err != nil {
			l.f = Filter{}
			l.err = l.p.Err
			return true
		}

		if !l.p.IP.IsLoopback() && !l.p.IP.IsUnspecified() {
			l.f = Filter{}
			l.err = &ResourceError{
				Line: l.p.Line,
				Err:  &HostsIPError{IP: l.p.IP},
			}
			return true
		}

		if len(l.p.Hosts) <= 0 {
			l.f = Filter{}
			l.err = &ResourceError{
				Line: l.p.Line,
				Err:  ErrMissingHostname,
			}
			return true
		}

		l.hs = l.p.Hosts
		l.setDomain(l.p.Hosts[0])
		return true
	}
	l.f = Filter{}
	l.err = l.p.Err
	return false
}

func (l *HostsLoader) setDomain(domain string) {
	domain, err := IDNAToASCII(domain)
	if err != nil {
		err = &ResourceError{
			Line: l.p.Line,
			Err:  err,
		}
	}

	l.f = Filter{Domain: domain}
	l.err = err
}

func (l *HostsLoader) Filter() Filter { return l.f }
func (l *HostsLoader) Err() error     { return l.err }

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
