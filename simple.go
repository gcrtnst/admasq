package main

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
