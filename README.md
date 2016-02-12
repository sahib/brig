# ``brig``: Ship your data around the world

![a somewhat gay brig](https://raw.githubusercontent.com/disorganizer/blog/master/static/img/brig.png)

[![go reportcard](https://goreportcard.com/badge/github.com/disorganizer/brig)](https://goreportcard.com/report/github.com/disorganizer/brig)
[![GoDoc](https://godoc.org/github.com/disorganizer/brig?status.svg)](https://godoc.org/github.com/disorganizer/brig)
[![Build Status](https://travis-ci.org/disorganizer/brig.svg?branch=master)](https://travis-ci.org/disorganizer/brig)

## Table of Contents

- [About](#about)
- [Installation](#installation)
- [Authors](#authors)

## About

``brig`` is a distributed & secure file synchronization tool (and more!)

This is a very early work in progress, so there are no details yet.
More information will follow once a rough first prototype is ready.
For now, you can [read this very chaotic blog](https://disorganizer.github.io/blog/).

Summarized in one paragraph, it is an ``syncthing``, ``git-annex`` or
``BTSync`` alternative, that gives you a commandline interface, a fuse
filesystem and a library that can encrypt and compress files which are in turn
distributed through ``ipfs`` while the file metadata is exchanged via XMPP.

Even shorter: It's supposed to be as flexible as ``git``, but for whole files.

## Installation

```bash
$ go get github.com/disorganizer/brig/brig
```

If that complains about some ``ipfs`` dependencies, you might need to follow the ``ipfs`` [install guide](https://github.com/ipfs/go-ipfs#build-from-source).

## Authors

| *Name*                                                 | *Active*   |
|--------------------------------------------------------|------------|
| Christopher <[sahib](https://github.com/sahib)> Pahl   | 2015-today |
| Christoph <[qitta](https://github.com/qitta)> Piechula | 2015-today |
