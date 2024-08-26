package main

import (
	"bytes"
	"net/netip"
)

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
