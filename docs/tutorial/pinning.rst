.. _pinning-section:

Pinning
-------

How can we control what files are stored locally and which should be retrieved
from the network? You can do this by **pinning** each file or directory you
want to keep locally. Normally, files that are not pinned may be cleaned up
from time to time, that means they are evaded from the local cache and need to
be fetched again when being accessed afterwards. Since you still have the
metadata for this file, you won't notice difference beside possible network
lag. When you pin a file, it will not be garbage collected.

``brig`` knows of two types of pins: **Explicit** and **implicit**.

- **Implicit pins:** This kind of pin is created automatically by ``brig`` and
  cannot be created by the user. In the command line output it always shows as
  blue pin. Implicit pins are created by ``brig`` whenever you create a new
  file, or update the contents of a file. The old version of a file will then
  be unpinned.
- **Explicit pins:** This kind of pin is created by the user explicitly (hence
  the name) and is never done by ``brig`` automatically. It has the same effect
  as an implicit pin, but cannot be removed again by ``brig``, unless
  explicitly unpinned. It's the user's way to tell ``brig`` »Never forgot
  these!«.

.. note::

    The current pinning implementation is still under conceptual development.
    It's still not clear what the best way is to modify/view the pin state
    of older versions. Time and user experience will tell.

When syncing with somebody, all files retrieved by them are by default **not
pinned**. If you want to keep them for longer, make sure to pin them
explicitly.

If you never pin something explicitly, only the newest version of all files
will be stored locally. If you decide that you need older versions, you can pin
them explicitly, so brig cannot unpin them implicitly. For this you should also
look into the ``brig pin set`` and ``brig pin clear`` commands, which are
similar to ``brig pin add`` and ``brig pin rm`` but can operate on whole commit
ranges.

Garbage collection
~~~~~~~~~~~~~~~~~~

Strongly related to pinning is garbage collection. This is normally being run
for you every few minutes, but you can also trigger it manually via the ``brig
gc`` command. While not usually needed, it can help you understand how ``brig``
works internally as it shows what hashes it throws away.

