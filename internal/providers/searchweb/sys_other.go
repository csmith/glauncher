//go:build !linux

package searchweb

import "syscall"

var openCommand = "open"

var syscallSetProcessGroupID = syscall.SysProcAttr{Setpgid: true}
