.. _quickstart:

Quickstart
==========

This does not really explain the philosophy behind ``brig``, but gives a good
idea what the tool is and how it's supposed to be used.

1. Init
-------

Before you can do anything with ``brig`` you need to create a repository.
During this step, also your online identity will be created. So make sure
to use a sane username (``sahib@wald.de``) and resource (``laptop``).

.. raw:: html

    <script src="https://asciinema.org/a/163713.js" id="asciicast-163713" async></script>

2. Adding files
---------------

Before synchronizing them, you need to *stage* them. The files will be stored
encrypted (and possibly compressed) in blobs on your harddisks.

.. raw:: html

    <script src="https://asciinema.org/a/j5yCj6H6fPUldbJDQz9nDhUc4.js" id="asciicast-j5yCj6H6fPUldbJDQz9nDhUc4" async></script>


3. Coreutils
------------

``brig`` provides implementations of most file related coreutils like ``mv``,
``cp``, ``rm``, ``mkdir`` or ``cat``. Handling of files should thus feel
familar for users that know the command line.

.. raw:: html

    <script src="https://asciinema.org/a/swIw29Qkml0A44H1MgKQvOQ8L.js" id="asciicast-swIw29Qkml0A44H1MgKQvOQ8L" async></script>

4. Mounting
-----------

For daily use and for use with other tools you might prefer a folder that contains the file
you gave to ``brig``. This can be done via the builtin FUSE layer.

.. raw:: html

    <script src="https://asciinema.org/a/166178.js" id="asciicast-166178" async></script>

.. note::

    Some built-in commands provided by brig are faster.
    ``brig cp`` for example only copies metadata, while the real ``cp`` will copy the whole file.

5. Commits
----------

In it's heart, ``brig`` is very similar to ``git`` and also supports versioning
via commits. In contrast to ``git`` however, there are no branches and you
can't go back in history -- you can only bring the history back up front.

.. raw:: html

    <script src="https://asciinema.org/a/166180.js" id="asciicast-166180" async></script>

6. History
----------

Each file (and directory) maintains a history of the operations that were done
to this file.


.. raw:: html

    <script src="https://asciinema.org/a/166181.js" id="asciicast-166181" async></script>

7. Discovery & Remotes
----------------------

In order to sync with your buddies, you need to add their *fingerprint* as remotes.
How do you get their fingerprint? In the best case by using a separate side channel
like telephone, encrypted email or elsewhise. But ``brig`` can assist finding remotes
via the ``brig net locate`` command.

.. raw:: html

    <script src="https://asciinema.org/a/166182.js" id="asciicast-166182" async></script>

.. note::

    You should **always** verify the fingerprint is really the one of your buddy.
    ``brig`` cannot do this for you.

8. Sync & Diff
--------------

Once both parties have setup each other as remotes, we can easily view and sync
with their data.

.. raw:: html

    <script src="https://asciinema.org/a/166183.js" id="asciicast-166183" async></script>

9. Pinning
----------

By default ``brig`` will only keep the most recent files. All other files will
be marked to deletions after a certain timeframe. This is done via *Pins*. If
a file is pinned, it won't get deleted. If you don't need a file in local
storage, you can also unpin it. On the next access ``brig`` will try to load it
again from a peer that provides it (if possible).

.. raw:: html

    <script src="https://asciinema.org/a/176590.js" id="asciicast-176590" async></script>
