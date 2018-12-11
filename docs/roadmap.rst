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

First Release
-------------

The first real release (0.2.0 »Baffling Buck«) is planned for end of November
2018. Until then, the software should provide the following basic features:

- Stable command line interface.
- Git-like version control
- User discovery
- User authentication
- Fuse filesystem

All of the above features are currently already implemented and work.
Focus is on stabilizing the features and making it somewhat release ready. All
those features combined do already provide some usefulness, but for being
a day-to-day useful tool, it takes a few more features, especially being able
to sync with offline peers over a trusted partner.

Note that there will be no stability guarantees before version 1.0.0.

Future
------

Those features should be considered after releasing the first prototype.
A certain amount of first user input should be collected to see if the
direction we're going is valid.

 ..  role:: strikethrough

*Gateway:* :strikethrough:`Provide a built-in (and optional) http server, that can »bridge«
between the internal ipfs network and people that use a regular browser.
Instances that run on a public server can then provide hyperlinks of files to
non-brig users.` *Done as of version 0.3.0.*

*Shelf instances:* Special instaces of brig, that operate automatically and are
meant to be run on public servers. They can be used to exchange data between
users that are not online at the same day (e.g. due to timezone differences).

*Automatic syncing:* Automatically publish changes after a short amount of time.
If an instance modified some file other nodes are notified and can decide to
pull the change.

*Intelligent pinning strategies:* By default only the most recent layer of
files are being kept. This is very basic and can't be configured currently.
Some users might only want to have only the last few used files pinned, archive
instances might want to pin almost everything up to a certain depth.

*Improve read/write performance:* Big files are currently hold in memory
completely by the fuse layer (when doing a flush). This is suboptimal and needs
more intelligent handling and out-of-memory caching of writes. Also, the
network performance is often very low and ridden by network errors and
timeouts.

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

*Implement a portable GUI:* Many user will rely on a GUI to configure brig and
hit the »sync button«. We should optionally provide this in a portable fashion
(browser based app? I kinda hate myself for proposing this though...).
Most of the time the GUI should be a simple tray icon that can be clicked to sync.
A rough and a little exaggerated mock-up was already drawn up for GNOME:

.. image:: _static/overview.svg
    :width: 66%
