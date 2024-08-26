package main

import (
	"bufio"
	"io"
)

type SimpleLoader struct {
	p   *SimpleParser
	exc bool

	f   Filter
	err error
}

func NewSimpleLoader(r io.Reader) *SimpleLoader {
	p := NewSimpleParser(r)
	return &SimpleLoader{p: p}
}

func (l *SimpleLoader) SetException(exc bool) {
	l.exc = exc
}

func (l *SimpleLoader) Load() bool {
	if !l.p.Parse() {
		l.f = Filter{}
		l.err = l.p.Err
		return false
	}

	var err error
	domain := l.p.Domain
	domain, err = IDNAToASCII(domain)
	if err != nil {
		err = &ResourceError{
			Line: l.p.Line,
			Err:  err,
		}
	}

	l.f = Filter{
		Exception: l.exc,
		Domain:    domain,
	}
	l.err = err
	return true
}

func (l *SimpleLoader) Filter() Filter { return l.f }
func (l *SimpleLoader) Err() error     { return l.err }

type SimpleParser struct {
	Line   int
	Domain string
	Err    error

	s    *bufio.Scanner
	lnum int
}

func NewSimpleParser(r io.Reader) *SimpleParser {
	s := bufio.NewScanner(r)
	s.Split(bufio.ScanLines)
	return &SimpleParser{s: s}
}

func (p *SimpleParser) Parse() bool {
	for p.s.Scan() {
		p.lnum++

		line := p.s.Bytes()
		domain := ParseSimpleLine(line)
		if domain != "" {
			p.Line = p.lnum
			p.Domain = domain
			p.Err = nil
			return true
		}
	}

	p.Line = p.lnum
	p.Domain = ""
	p.Err = p.s.Err()
	return false
}

func ParseSimpleLine(line []byte) string {
	lo := 0
	for ; lo < len(line) && (line[lo] == ' ' || line[lo] == '\t'); lo++ {
	}

	hi := lo
	for i := lo; i < len(line) && line[i] != '#'; i++ {
		if line[i] != ' ' && line[i] != '\t' {
			hi = i + 1
		}
	}

	return string(line[lo:hi])
}
