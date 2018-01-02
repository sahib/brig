# ``brig``: Ship your data around the world

<center>  <!-- I know, that's not how you usually do it :) -->
<img src="https://raw.githubusercontent.com/sahib/brig/master/docs/logo.png" alt="a brig" width="50%">
</center>

[![go reportcard](https://goreportcard.com/badge/github.com/sahib/brig)](https://goreportcard.com/report/github.com/sahib/brig)
[![GoDoc](https://godoc.org/github.com/sahib/brig?status.svg)](https://godoc.org/github.com/sahib/brig)
[![Build Status](https://travis-ci.org/sahib/brig.svg?branch=master)](https://travis-ci.org/sahib/brig)
[![Documentation](https://readthedocs.org/projects/rmlint/badge/?version=latest)](http://brig.readthedocs.io/en/latest)

## Table of Contents

- [About](#about)
- [Installation](#installation)
- [Authors](#authors)

## About

``brig`` is a distributed & secure file synchronization tool with version control.
It is based on ``ipfs``, written in Go and will feel familiar to ``git`` users.

Key feature highlights:
* Works even for nodes that are hidden behind a NAT.
* Encryption of data in rest and transport + compression on the fly.
* Simplified ``git`` version control (no real branches).
* Sync algorithm that can handle moved files and empty directories and files.
* Your data does not need to be stored on the device you are using.
* FUSE filesystem that feels like a normal (sync) folder.
* No central server at all. Still, central architectures can be build with ``brig``.
* Simple user management with users that look like email addresses.
* Hash algorithm can be changed, unlike with ``git``. ;-)

----

This project has started end of 2015 and has seen many conceptual changes in
the meantime. It started out as research project of two computer science
students (me and [qitta](https://github.com/qitta)). After writing our [master
theses](https://github.com/sahib/brig-thesis) on it, it was put down for
a few months until I ([sahib](https://github.com/sahib)) picked at up again and
currently am trying to push it to a usable prototype.

### Donations

In it's current status, it's a working proof of concept. I'd love to work on it
more, but my day job (and the money that comes with it) forbids that.
If you're interested in the development of ``brig`` and would think about
supporting me financially, then please [contact me!](mailto:sahib@online.de)

If you'd like to give me a small donation, you can use *liberapay*:

<noscript><a href="https://liberapay.com/sahib/donate"><img alt="Donate using Liberapay" src="https://liberapay.com/assets/widgets/donate.svg"></a></noscript>

### Focus

``brig`` tries to focus on being up conceptually simple, by hiding a lot of
complicated details regarding storage and security. Therefore I hope the end
result is easy and pleasant to use, while being to be secure by default.
Since ``brig`` is a "general purpose" tool for file synchronization it of course
cannot excel in all areas. This is especially true for efficiency, which is
sometimes sacrificed to get the balance of usability and security right.

## Installation

```bash
$ go get github.com/sahib/brig/cmd/brig
```

That should just work if you previously [setup Go](https://golang.org/doc/install).
Afterwards you'll have a ``brig`` command on your computer, which will print it's help when invoked without any
arguments.

## Getting started

TODO: Make this an asciinema.

```bash
$ mkdir sync
$ cd sync
$ brig init alice@wonderland.de
$ brig cat README.md
$ brig remote add bob@wonderland.de QM123...:Smxyz...
```
