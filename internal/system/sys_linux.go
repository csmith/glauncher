package system

import "syscall"

var OpenCommand = "xdg-open"

var processGroupAttr = syscall.SysProcAttr{Setpgid: true}
