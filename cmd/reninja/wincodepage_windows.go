//go:build windows

package main

import "golang.org/x/sys/windows"

const cpUTF8 = 65001

func buildFileEncoding() string {
	if windows.GetACP() == cpUTF8 {
		return "UTF-8"
	}
	return "ANSI"
}
