package main

func cRed(s string) string {
	return "\x1b[31m" + s + "\x1b[39m"
}

func cGrn(s string) string {
	return "\x1b[32m" + s + "\x1b[39m"
}

func cYel(s string) string {
	return "\x1b[33m" + s + "\x1b[39m"
}

func cBlu(s string) string {
	return "\x1b[34m" + s + "\x1b[39m"
}
