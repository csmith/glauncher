//go:build !linux

package desktop

import "syscall"

var syscallSetProcessGroupID = syscall.SysProcAttr{}
