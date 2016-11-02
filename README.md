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
``Resilio``-alternative, that gives you a commandline interface, a fuse
filesystem and a library that can encrypt and compress files which are in turn
distributed through ``ipfs`` while the file metadata is transmitted separetely.
It's a bit similar to the currently also unfinished [bazil](https://bazil.org) maybe.

Even shorter: It's supposed to be as flexible as ``git``, but for complete files.

A master thesis on ``brig`` [has been written](https://github.com/disorganizer/brig-thesis), 
which is only available in german though. That doesn't mean we're planning to
discontinue it after that thesis - actually we'd love to get paid for
developement! Care to throw money at us?

## Installation

```bash
$ go get github.com/disorganizer/brig/brig
```

That should just work if you previously [setup
Go](https://golang.org/doc/install). Afterwards you'll have a ``brig`` command
on your computer, which will print it's help when invoked without any
arguments.

## Authors

| *Name*                                                 | *Active*   |
|--------------------------------------------------------|------------|
| Christopher <[sahib](https://github.com/sahib)> Pahl   | 2015-today |
| Christoph <[qitta](https://github.com/qitta)> Piechula | 2015-today |
