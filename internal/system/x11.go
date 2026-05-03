package system

/*
#cgo pkg-config: x11

#include <X11/Xlib.h>
#include <stdlib.h>
*/
import "C"

import "unsafe"

func RemoveX11Decorations(display unsafe.Pointer, window uintptr) {
	dpy := (*C.Display)(display)
	win := C.Window(window)

	name := C.CString("_MOTIF_WM_HINTS")
	defer C.free(unsafe.Pointer(name))
	atom := C.XInternAtom(dpy, name, C.True)
	if atom == 0 {
		return
	}

	var hints [5]C.long
	hints[0] = 2
	hints[2] = 0

	C.XChangeProperty(dpy, win, atom, atom, 32,
		C.PropModeReplace, (*C.uchar)(unsafe.Pointer(&hints[0])), 5)
	C.XFlush(dpy)
}
