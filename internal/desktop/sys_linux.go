package desktop

import "syscall"

var syscallSetProcessGroupID = syscall.SysProcAttr{Setpgid: true}
