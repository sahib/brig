---
date:        "2016-01-15T01:13:07-07:00"
title:       "A small step for mankind, a big step for brig"
description: "First devlog entry"
tags:        [ "Development", "Go", "brig"]
topics:      [ "Development", "Go" ]
slug:        "devlog"
---

A small historic moment was achieved today: The very first file was added to
``brig``. There was no way to get it out again, but hey - Progress comes in
steps. Luckily, just two hours later there was a ``brig get`` command that
could retrieve the file again from ``ipfs``.

This is also my very first devlog entry, so... Hi. I mainly write this to
remember what I did (and when) on the course of the project. Also it sometimes
is really useful to reflect on what kind of boolshit I wrote today. Ever
noticed that you get the best ideas doing arbitrary things like peeing? That's
the same effect, I guess. If it's fun to read for others... that's okay too.
I try to keep it updated after every more or less productive session.
That might mean daily, that might also mean once a week.

So, back to the technical side of life. ``brig add`` currently works a bit
confusing. It is supposed to read a regular file on the disk, compress and
encrypt it and add it to ``ipfs``. The encryption and compression layer uses
``io.Writer`` though, so we can't just stack ``io.Reader`` on top of each
other. Instead we need to use a nice little feature from the stdlib:
``io.Pipe()``. This function returns a ``io.Writer`` and a ``io.Reader``. Every
write on the writer produces a corresponding read on the reader - without internal
copying of the data. Yay. If you have a piece of API that needs a ``io.Reader``,
but you just have a ``io.Writer``, then ``io.Pipe()`` should pop into your mind now.

Here's how it looks in practice:

```go
func NewFileReader(key []byte, r io.Reader) (io.Reader, error) {
	pr, pw := io.Pipe()

	// Setup the writer part:
	wEnc, err := encrypt.NewWriter(pw, key)
	if err != nil {
		return nil, err
	}

	wZip := compress.NewWriter(wEnc)

	// Suck the reader empty and move it to `wZip`.
	// Every write to wZip will be available as read in `pr`.
	go func() {
		defer func() {
			wEnc.Close()
			pw.Close()
		}()

		if _, err := io.Copy(wZip, r); err != nil {
			// TODO: Warn or pass to outside?
			log.Warningf("add: copy: %v", err)
		}
	}()

	return pr, nil
}
```

That's all for today! For tomorrow a cleanup session is planned and the piece
of code that derives the AES-Key from an unencrypted file.
