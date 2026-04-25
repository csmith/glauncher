//go:build !linux

package folders

import "syscall"

var openCommand = "open"

var syscallSetProcessGroupID = syscall.SysProcAttr{}
