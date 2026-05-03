package code

import "syscall"

var syscallSetProcessGroupID = syscall.SysProcAttr{Setpgid: true}
