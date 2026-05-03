package system

import (
	execcmd "os/exec"
)

func Launch(name string, args ...string) error {
	c := execcmd.Command(name, args...)
	c.Stdin = nil
	c.Stdout = nil
	c.Stderr = nil
	c.SysProcAttr = &processGroupAttr
	return c.Start()
}

func OpenURL(url string) error {
	return Launch(OpenCommand, url)
}
