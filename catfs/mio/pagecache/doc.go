// Package overlay implements a io.ReaderAt and io.WriterAt that is similar in
// function to the OverlayFS of Linux. It overlays a read-only stream and
// enables write support. The writes will take priority on the data in stream
// and will therefore be visible when calling ReadAt() of the overlay.
// Read() and Write() are currently not supported, since they would not be used
// by brig.
//
// Note that the normal POSIX file operations are supported. This includes
// truncating a file to a certain length and also extending it to a certain
// length. If length of the overlay is greater than the size of the underlying
// stream we pad it with zeros - just like the kernel would do. Files can be
// also extended by writing new blocks to the end of the overlay.
//
// Seeking will be done when necessary. WriteAt() has to do no seeking at all,
// while ReadAt() will only seek if it has to (i.e. not reading from cache
// alone, or if we're not if the right offset already).
//
// Implementation detail: The stream is divided into same-sized pages. Each
// page can be retrieved as whole from the cache. If a page for a certain read
// offset is found, then ReadAt() will overlay it with the underlying stream or
// even read from memory if the stream completely occludes the underlying
// stream. In general, care was taken to optimize a bit more for Write() since
// pages delivered by ReadAt() can be cached by the FUSE filesystem.
//
// You can choose the page cache when creating the overlay. Depending on the
// page cache implementation it's also possible to edit large files and
// make edits persistent.
//
// NOTE: Whenever int32 is used in this code, it refers to per-page offsets or
// size. When int64 is used the content is an offset of the underlying offset.
package overlay
