Roadmap
=======

This document lists the improvements that can be done to ``brig`` and (if
possible) when. All features below are not guaranteed to be implemented and,
can be seen more as possible improvements that might change during
implementation. Also it should be noted that each future is only an idea and
not a fleshed ou implementation plan.

Bug fixes and minor changes in handling are not included since this document is
only for »big picture« ideas. Also excluded are stability/performance
improvements, documentation and testing work, since this is part of the
»normal« development.

Current state
-------------

The first real release (0.3.0 »Galloping Galapagos«) was released on the 7th December 2018.
It includes all basic features and is working somewhat. The original goals were met:

- Stable command line interface.
- Git-like version control
- User discovery
- User authentication
- Fuse filesystem

For day-to-day use there are quite some other features that make brig easier to use
and capable of forming a Dropbox-like backend out of several nodes.

**There will be no stability guarantees before version 1.0.0.**

Future
------

Those features should be considered after releasing the first prototype.
A certain amount of first user input should be collected to see if the
direction we're going is valid.

 ..  role:: strikethrough

**Gateway:** :strikethrough:`Provide a built-in (and optional) http server, that can »bridge«
between the internal ipfs network and people that use a regular browser.
Instances that run on a public server can then provide hyperlinks of files to
non-brig users.` *Done as of version 0.3.0.*

**Config profiles:** Make it easy to configure brig in a way to serve either as thin client
or as archival node. Archival nodes can be used in cases where a brig network spans over computers
that lie in a different timezone. The archival node would accumulate all changes and repositories
would see it as some sort of "blessed repository" which holds the latest and greatest state.

**Automatic syncing:** :strikethrough:`Automatically publish changes after a short amount of time.
If an instance modified some file other nodes are notified and can decide to
pull the change.` *Done as of version 0.4.0.*

**Intelligent pinning strategies:** :strikethrough:`By default only the most recent layer of
files are being kept. This is very basic and can't be configured currently.
Some users might only want to have only the last few used files pinned, archive
instances might want to pin almost everything up to a certain depth.` *Done as of version 0.4.0 (see repinning)*

*Improve read/write performance:* Big files are currently hold in memory
completely by the fuse layer (when doing a flush). This is suboptimal and needs
more intelligent handling and out-of-memory caching of writes. Also, the
network performance is often very low and ridden by network errors and
timeouts. This can be tackled since IPFS v0.4.19 supports an --offline switch to
error out early if a file is not available locally.

*More automated authentication scheme:* E-Mail-like usernames could be used to
verify a user without exchanging fingerprints. This could be done by e.g.
sending an activation code to the email of an user (assuming the brig name is
the same as his email), which the brig daemon on his side could read and send back.

*Format and read fingerprint from QR-Code:* Fingerprints are hard to read and
not so easy to transfer and verify. QR-Code could be a solution here, since we
could easily snap a picture with a phone camera or print it on a business card.

Far Future
----------

Those features are also important, but require some more in-depth research or
more work and are not the highest priority currently.

*Port to other platforms:* Especially Windows and eventually Android. This
relies on external help, since I'm neither capable of porting it, nor really
a fan of both operating systems.

*Implement alternative to fuse:* FUSE currently only works on Linux and is
therefore not usable outside of that. Windows has something similar (called
Dokan_). Alternatively we could also go on by implementing a WebDAV server,
which can also be mounted.

.. _dokan: https://github.com/keybase/kbfs/tree/master/dokan

*Ensure N-Copies:* It should be possible to define a minimum amount of copies
a file has to have on different peers. This could be maybe incorporated into
the pinning concept. If a user wants to remove a file, brig should warn him if
he would violate the min-copies rule. This idea is shamelessly stolen from
``git-annex``.
