//go:build !linux

package code

import "syscall"

var syscallSetProcessGroupID = syscall.SysProcAttr{}
