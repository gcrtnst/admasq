package main

import (
	"bufio"
	"io"
)

type SimpleParser struct {
	s *bufio.Scanner

	lnum   int
	domain string
	err    error
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
			p.domain = domain
			p.err = nil
			return true
		}
	}

	p.domain = ""
	p.err = p.s.Err()
	return false
}

func (p *SimpleParser) Line() int      { return p.lnum }
func (p *SimpleParser) Domain() string { return p.domain }
func (p *SimpleParser) Err() error     { return p.err }

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
