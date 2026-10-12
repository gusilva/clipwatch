import Foundation
import ServiceManagement

/// The login item that runs the clipboard watcher.
///
/// The job is described by
/// `Clipwatch.app/Contents/Library/LaunchAgents/gusilva.clipwatch.watcher.plist`.
///
/// - Note: `SMAppService` finds that file through `Bundle.main`, so the
///   functions below work only when the running binary is
///   `Clipwatch.app/Contents/MacOS/clipwatch`.
private let watcher = SMAppService.agent(plistName: "gusilva.clipwatch.watcher.plist")

/// Returns the registration status of the watcher.
///
/// - Returns: `0` not registered, `1` enabled, `2` requires approval,
///   `3` not found.
@_cdecl("smStatus")
public func smStatus() -> Int32 {
    return Int32(watcher.status.rawValue)
}

/// Registers the watcher with the system.
///
/// - Returns: `NULL` on success. On failure, a message allocated with
///   `malloc` that the caller must free.
@_cdecl("smRegister")
public func smRegister() -> UnsafeMutablePointer<CChar>? {
    do { try watcher.register(); return nil } catch { return strdup("\(error)") }
}

/// Stops the watcher and removes its registration.
///
/// - Returns: `NULL` on success. On failure, a message allocated with
///   `malloc` that the caller must free.
@_cdecl("smUnregister")
public func smUnregister() -> UnsafeMutablePointer<CChar>? {
    do { try watcher.unregister(); return nil } catch { return strdup("\(error)") }
}

/// Opens System Settings at General > Login Items & Extensions.
@_cdecl("smOpenSettings")
public func smOpenSettings() {
    SMAppService.openSystemSettingsLoginItems()
}
