Roadmap
=======

This document is currently more a collection of notes to me, i.e. what features
to implement and when. It could however be useful or interesting for others to
see what's on my todo list (and thus, what the current state is). This does not
include bug fixes or minor features, only the big picture ideas.

First Release
-------------

The first real relaese (i.e. show it to people) is planned for end of April 2018.
Until then, the software should provide the following:

- Stable cli interface.
- Git-like version control
- User discovery
- User authentication
- Fuse filesystem

Most of the above features are currently already implemented and somewhat work.
Focus is on stabilizing the features and making it somewhat release ready.
All those features combined do already provide some usefulness, but
for being a day-to-day useful tool, it takes a bit more...

Future
------

*Partial diffs:* Currently the whole store is being sent on every fetch.
Clients should be able to only request (and provide) the diff between
two commits.

*Gateway:* Provide a built-in (and optional) http server, that can »bridge«
between the internal ipfs network and people that use a regular browser.
Instances that run on a public server can then provide hyperlinks of files to
non-brig users.

*Shelf instances:* Special instaces of brig, that operate automatically and are
meant to be run on public servers. They can be used to exchange data between
users that are not online at the same day (e.g. due to timezone differences).

*Automatic syncing:* Automatically publish changes after a short amount of time.
If an instance modified some file other nodes are notified and can decide to
pull the change. This relies on *partial diffs*.

*Intelligent pinning strategies:* By default only the most recent layer of files
are being kept. This is very basic and can't be configured. Some users might only
want to have only the last few used files pinned, archive instances might want
to pin almost everything up to a certain depth.

*Improve read/write performance:* Big files are currently hold in memory
completely by the fuse layer (when doing a flush). This is suboptimal and needs
more intelligent handling and out-of-memory caching of writes.

*More automated authentication scheme:* EMail-like usernames could be used to
verify a user without exchanging fingerprints. This could be done by e.g.
sending an activation code to the email of an user (assuming the brig name is
the same as his email), which the brig daemon on his side could read and send back.

*Format and read fingerprint from QRCode:* Fingerprints are hard to read and
not so easy to transfer and verify. QRCode could be a solution here, since we
could easily snap a picture with a phone camera.

Far Future
----------

*Port to other platforms:* Especially Windows and eventually Android. This
relies on external help, since I'm neither capable of porting it, nor really
a fan of both operating systems.

*Implement alternative to fuse:* FUSE currently only works on Linux and is
therefore not usable outside of that. Windows has something similar (called
Dokan: https://github.com/keybase/kbfs/tree/master/dokan). Alternatively we
could also go on by implementing a WebDAV server, which can also be mounted.

*Implement a portable GUI:* Many user will rely on a GUI to configure brig and
hit the »sync button«. We should optionally provide this in a portable fashion
(browser based app? I kinda hate myself for proposing this though...)
