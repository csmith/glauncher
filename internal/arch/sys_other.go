//go:build !linux

package arch

import "syscall"

var openCommand = "open"

var syscallSetProcessGroupID = syscall.SysProcAttr{Setpgid: true}
