# `brig`: Ship your data around the world

<center>  <!-- I know, that's not how you usually do it :) -->
<img src="https://raw.githubusercontent.com/sahib/brig/master/docs/logo.png" alt="a brig" width="50%">
</center>

[![go reportcard](https://goreportcard.com/badge/github.com/sahib/brig)](https://goreportcard.com/report/github.com/sahib/brig)
[![GoDoc](https://godoc.org/github.com/sahib/brig?status.svg)](https://godoc.org/github.com/sahib/brig)
[![Build Status](https://travis-ci.org/sahib/brig.svg?branch=master)](https://travis-ci.org/sahib/brig)
[![Documentation](https://readthedocs.org/projects/rmlint/badge/?version=latest)](http://brig.readthedocs.io/en/latest)
[![License: AGPL v3](https://img.shields.io/badge/License-AGPL%20v3-blue.svg)](https://www.gnu.org/licenses/agpl-3.0)
[![CII Best Practices](https://bestpractices.coreinfrastructure.org/projects/1558/badge)](https://bestpractices.coreinfrastructure.org/en/projects/1558)

![](https://raw.githubusercontent.com/sahib/brig/master/docs/_static/gateway-files.png)

## Table of Contents

- [`brig`: Ship your data around the world](#brig-ship-your-data-around-the-world)
  - [Table of Contents](#table-of-contents)
  - [About](#about)
  - [Installation](#installation)
  - [Getting started](#getting-started)
  - [Status](#status)
  - [Documentation](#documentation)
  - [Donations](#donations)
  - [Focus](#focus)

## About

`brig` is a distributed & secure file synchronization tool with version control.
It is based on `IPFS`, written in Go and will feel familiar to `git` users.

**Key feature highlights:**

* Encryption of data in rest and transport + compression on the fly.
* Simplified `git` version control.
* Sync algorithm that can handle moved files and empty directories and files.
* Your data does not need to be stored on the device you are currently using.
* FUSE filesystem that feels like a normal (sync) folder.
* No central server at all. Still, central architectures can be build with `brig`.
* Simple user identification and discovery with users that look like email addresses.

Also take a look [at the documentation](http://brig.readthedocs.io/en/latest/index.html) for more details.

## Installation

If you have a recent version of `go` (`>= 1.9`) installed, it should be as easy as this:

```bash
$ go get -d -v -u github.com/sahib/brig  # Download the sources.
$ cd $GOPATH/src/github.com/sahib/brig   # Go to the source directory.
$ make                                   # Build the software.
$ $GOPATH/bin/brig help                  # Run the binary
```

Please refer to the [install docs](https://brig.readthedocs.io/en/latest/installation.html) for more details.

## Getting started

[![asciicast](https://asciinema.org/a/163713.png)](https://asciinema.org/a/163713)

...If you want to know, what to do after you can read the
[Quickstart](http://brig.readthedocs.io/en/latest/quickstart.html).

## Status

This software is in a **beta phase** currently. All mentioned features should
work. Things might still change rapdily and there will be no guarantee to
change that until version `1.0.0`. Do not use `brig` yet as only storage
for your production data. There are still bugs, but it should be safe enough to
toy around with it quite a bit.

This project has started end of 2015 and has seen many conceptual changes in
the meantime. It started out as research project of two computer science
students (me and [qitta](https://github.com/qitta)). After writing our [master
theses](https://github.com/disorganizer/brig-thesis) on it, it was put down for
a few months until I ([sahib](https://github.com/sahib)) picked at up again and
currently am trying to push it to usable software.

If you want to open a bug report, just type `brig bug` to get a readily filled template for you.

## Documentation

All documentation can be found on [ReadTheDocs.org](http://brig.readthedocs.io/en/latest/index.html).

## Donations

I really would like to work more on `brig`, but my day job (and the money
that comes with it) forbids that. If you're interested in the development and
would think about supporting me financially, then please [contact
me!](mailto:sahib@online.de) If you'd like to give me a small & steady
donation, you can always use *Liberapay*:

<noscript><a href="https://liberapay.com/sahib/donate"><img alt="Donate using Liberapay" src="https://liberapay.com/assets/widgets/donate.svg"></a></noscript>

*Thank you!*

## Focus

`brig` tries to focus on being up conceptually simple, by hiding a lot of
complicated details regarding storage and security. Therefore I hope the end
result is easy and pleasant to use, while being secure by default.
Since `brig` is a "general purpose" tool for file synchronization it of course
cannot excel in all areas. This is especially true for efficiency, which is
sometimes sacrificed to get the balance of usability and security right.
