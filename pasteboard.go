package main

/*
#cgo LDFLAGS: -L${SRCDIR}/build -lpasteboard -framework AppKit -framework Foundation -L/usr/lib/swift -Wl,-rpath,/usr/lib/swift
#include <stdlib.h>

// Implemented in swift/pasteboard.swift (@_cdecl), linked from build/libpasteboard.a.
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

func changeCount() int {
	return int(C.pbChangeCount())
}

func hasType(uti string) bool {
	cs := C.CString(uti)
	defer C.free(unsafe.Pointer(cs))
	return C.pbHasType(cs) == 1
}

func readString() (string, bool) {
	p := C.pbString(C.long(MaxEntryBytes))
	if p == nil {
		return "", false // clipboard holds no text
	}

	defer C.free(unsafe.Pointer(p))
	return C.GoString(p), true
}

func writeString(s string) (int, error) {
	cs := C.CString(s)
	defer C.free(unsafe.Pointer(cs))

	n := C.pbWriteString(cs)

	if n < 0 {
		return 0, errors.New("pasteboard write failed")
	}

	return int(n), nil
}
