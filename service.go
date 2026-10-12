package main

/*
#cgo LDFLAGS: -framework ServiceManagement
#include <stdlib.h>

int smStatus(void);
char *smRegister(void);
char *smUnregister(void);
void smOpenSettings(void);
*/
import "C"
import (
	"errors"
	"unsafe"
)

// serviceStatus is the registration state of the watcher. Its values mirror
// SMAppService.Status.
type serviceStatus int

const (
	notRegistered    serviceStatus = 0
	enabled          serviceStatus = 1
	requiresApproval serviceStatus = 2
	notFound         serviceStatus = 3
)

// String returns the name of the status.
func (s serviceStatus) String() string {
	return [...]string{"notRegistered", "enabled", "requiresApproval", "notFound"}[s]
}

// watcherStatus returns the current registration state of the watcher.
func watcherStatus() serviceStatus {
	return serviceStatus(C.smStatus())
}

// registerWatcher registers the watcher with macOS. It fails unless the binary
// runs from inside Clipwatch.app.
func registerWatcher() error { return smErr(C.smRegister()) }

// unregisterWatcher stops the watcher and removes its registration.
func unregisterWatcher() error { return smErr(C.smUnregister()) }

// openLoginItems opens System Settings at Login Items & Extensions.
func openLoginItems() { C.smOpenSettings() }

// smErr converts the result of a Swift call into an error. It returns nil for
// a NULL message, and otherwise returns an error with the message, which it
// frees.
func smErr(cmsg *C.char) error {
	if cmsg == nil {
		return nil
	}

	defer C.free(unsafe.Pointer(cmsg))
	return errors.New(C.GoString(cmsg))
}
