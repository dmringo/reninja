//go:build !windows

package main

func buildFileEncoding() string {
	return "ANSI"
}
