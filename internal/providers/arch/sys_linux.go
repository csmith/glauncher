package arch

import "syscall"

var openCommand = "xdg-open"

var syscallSetProcessGroupID = syscall.SysProcAttr{Setpgid: true}
