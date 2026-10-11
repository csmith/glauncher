package system

import (
	"context"
	"errors"
	execcmd "os/exec"
	"strings"
	"syscall"
)

var OpenCommand = "xdg-open"

var processGroupAttr = syscall.SysProcAttr{Setpgid: true}

func Launch(name string, args ...string) error {
	c := execcmd.CommandContext(context.Background(), name, args...)
	c.Stdin = nil
	c.Stdout = nil
	c.Stderr = nil
	c.SysProcAttr = &processGroupAttr
	return c.Start()
}

func OpenURL(url string) error {
	return Launch(OpenCommand, url)
}

func CopyToClipboard(text string) error {
	var name string
	var args []string

	if p, _ := execcmd.LookPath("wl-copy"); p != "" {
		name = "wl-copy"
	} else if p, _ := execcmd.LookPath("xclip"); p != "" {
		name = "xclip"
		args = []string{"-selection", "clipboard"}
	} else if p, _ := execcmd.LookPath("xsel"); p != "" {
		name = "xsel"
		args = []string{"--clipboard", "--input"}
	} else {
		return errors.New("no clipboard utility found (install wl-copy, xclip, or xsel)")
	}

	cmd := execcmd.CommandContext(context.Background(), name, args...)
	cmd.Stdin = strings.NewReader(text)
	return cmd.Run()
}
