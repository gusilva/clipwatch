import AppKit

/// Returns the change count of the general pasteboard.
///
/// The count grows each time the pasteboard content changes.
@_cdecl("pbChangeCount")
public func pbChangeCount() -> Int {
    return NSPasteboard.general.changeCount
}

/// Reports whether the general pasteboard holds an item of the given type.
///
/// - Parameter uti: The pasteboard type identifier, as a C string.
/// - Returns: `1` if the type is present, otherwise `0`.
@_cdecl("pbHasType")
public func pbHasType(_ uti: UnsafePointer<CChar>) -> Int32 {
    let t = NSPasteboard.PasteboardType(String(cString: uti))
    return NSPasteboard.general.types?.contains(t) == true ? 1 : 0
}

/// Reads the text on the general pasteboard.
///
/// Text over `maxBytes` is never copied out of Swift.
///
/// - Parameter maxBytes: The largest UTF-8 size to accept.
/// - Returns: A string allocated with `malloc` that the caller must free, or
///   `NULL` if there is no text or it is too large.
@_cdecl("pbString")
public func pbString(_ maxBytes: Int) -> UnsafeMutablePointer<CChar>? {
    guard let s = NSPasteboard.general.string(forType: .string) else { return nil }
    if s.utf8.count > maxBytes { return nil }
    return strdup(s)
}

/// Replaces the general pasteboard content with a string.
///
/// The item is tagged as placed by a history app (see nspasteboard.org), with
/// an empty source because the original source is unknown.
///
/// - Parameter s: The new content, as a UTF-8 C string.
/// - Returns: The new change count, or `-1` if `s` is not valid UTF-8 or the
///   write fails.
@_cdecl("pbWriteString")
public func pbWriteString(_ s: UnsafePointer<CChar>) -> Int {
    let pb = NSPasteboard.general
    guard let str = String(validatingUTF8: s) else { return -1 }
    let n = pb.clearContents()
    if !pb.setString(str, forType: .string) { return -1 }
    pb.setString("", forType: NSPasteboard.PasteboardType("org.nspasteboard.source"))
    return n
}
