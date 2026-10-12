import AppKit

@_cdecl("pbChangeCount")
public func pbChangeCount() -> Int {
    return NSPasteboard.general.changeCount
}

@_cdecl("pbHasType")
public func pbHasType(_ uti: UnsafePointer<CChar>) -> Int32 {
    let t = NSPasteboard.PasteboardType(String(cString: uti))
    return NSPasteboard.general.types?.contains(t) == true ? 1 : 0
}

// Returns a malloc'd C string (caller frees), or NULL if there is no text or
// it exceeds maxBytes (too big: never copied into Go).
@_cdecl("pbString")
public func pbString(_ maxBytes: Int) -> UnsafeMutablePointer<CChar>? {
    guard let s = NSPasteboard.general.string(forType: .string) else { return nil }
    if s.utf8.count > maxBytes { return nil }
    return strdup(s)
}

// Replaces the clipboard with s, tagged as re-placed by a history app
// (nspasteboard.org). Returns the new changeCount, or -1.
@_cdecl("pbWriteString")
public func pbWriteString(_ s: UnsafePointer<CChar>) -> Int {
    let pb = NSPasteboard.general
    guard let str = String(validatingUTF8: s) else { return -1 }  // not valid UTF-8
    let n = pb.clearContents()                                    // take ownership; bumps changeCount
    if !pb.setString(str, forType: .string) { return -1 }
    pb.setString("", forType: NSPasteboard.PasteboardType("org.nspasteboard.source"))
    return n
}
