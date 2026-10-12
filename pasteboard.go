package main

/*
#cgo LDFLAGS: -L${SRCDIR}/build -lpasteboard -framework AppKit -framework Foundation -L/usr/lib/swift -Wl,-rpath,/usr/lib/swift
#include <stdlib.h>

long pbChangeCount(void);
int pbHasType(const char *util);
char *pbString(long maxBytes);
long pbWriteString(const char *s);
*/
import "C"
import (
	"errors"
	"unsafe"
)

// changeCount returns the change count of the general pasteboard, which grows
// each time its content changes.
func changeCount() int {
	return int(C.pbChangeCount())
}

// hasType reports whether the pasteboard holds an item of the given type.
func hasType(uti string) bool {
	cs := C.CString(uti)
	defer C.free(unsafe.Pointer(cs))
	return C.pbHasType(cs) == 1
}

// readString returns the text on the pasteboard. It returns false when there
// is no text or the text is larger than MaxEntryBytes.
func readString() (string, bool) {
	p := C.pbString(C.long(MaxEntryBytes))
	if p == nil {
		return "", false
	}

	defer C.free(unsafe.Pointer(p))
	return C.GoString(p), true
}

// writeString replaces the pasteboard content with s, marked as placed by a
// history app, and returns the new change count.
func writeString(s string) (int, error) {
	cs := C.CString(s)
	defer C.free(unsafe.Pointer(cs))

	n := C.pbWriteString(cs)

	if n < 0 {
		return 0, errors.New("pasteboard write failed")
	}

	return int(n), nil
}
