//go:build !linux

package system

import "syscall"

var OpenCommand = "open"

var processGroupAttr = syscall.SysProcAttr{}
