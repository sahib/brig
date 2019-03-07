.. _pinning-section:

Pinning
-------

How can we control what files are stored locally and which should be retrieved
from the network? You can do this by **pinning** each file or directory you
want to keep locally. Normally, files that are not pinned may be cleaned up
from time to time, that means they are evaded from the local cache and need to
be fetched again from the network when being accessed again. Since you still
have the metadata for this file, you won't notice the difference beside some
possible network lag. When you pin a file however, it will not be garbage
collected and stays in your local cache until unpinned.

``brig`` knows of two types of pins: **Explicit** and **implicit**.

- **Implicit pins:** This kind of pin is created automatically by ``brig`` and
  cannot be created by the user. In the command line output it is always shows as
  blue pin. Implicit pins are created by ``brig`` whenever you create a new
  file, or update the contents of a file. Implicit pins are managed by ``brig`` and
  as you will see later, it might decide to save you some space by unpinning old versions.
- **Explicit pins:** This kind of pin is created by the user explicitly (hence
  the name) and is never done by ``brig`` automatically. It has the same effect
  as an implicit pin, but cannot be removed again by ``brig``, unless
  explicitly unpinned by the user. This is a good way of telling ``brig`` to
  never unpin this specific version. Use this with care, since it is easy to forget about
  explicit pins.

When syncing with somebody, all files retrieved by them are by default **not
pinned**. If you want to keep them for longer, make sure to pin them
explicitly.

Garbage collection
~~~~~~~~~~~~~~~~~~

Strongly related to pinning is garbage collection. Whenever you need to clean up some
space, you can just type ``brig gc`` to remove all unpinned files from the cache.

By default, the garbage collector is also run once every hour. You can change this interval
by setting ``brig config set repo.autogc.interval`` to ``30m`` for example. You can also disable
this automatic garbage collection by issuing ``brig config set repo.autogc.enabled false``.

Repinning
~~~~~~~~~

Repinning allows you to control how many versions of each file you want to
store and/or how much space you want to store at most. The repinning feature is
controlled by the following configuration variables:

- **fs.repin.quota**: Maximum amount of data to store in a repository.
- **fs.repin.min_depth**: Keep this many versions definitely pinned. Trumps quota.
- **fs.repin.max_depth**: Unpin versions beyond this depth definitely. Trumps quota.
- **fs.repin.enabled**: Wether we should allow the repinning to run at all.
- **fs.repin.interval**: How much time to wait between calling repinning automatically.

Normally repinning will run for you every 15 minutes. You can also trigger it manually:

.. code-block:: bash

   $ brig pin repin

By default, ``brig`` will keep 1 version definitely (**fs.repin.min_depth**)
and delete all versions starting with the 10th (**fs.repin.max_depth**). The
default quota (**fs.repin.quota**) is 5GB. If repin detects files that need to
be unpinned, then it will first unpin all files that are beyond the max depth
setting. If this is not sufficient to stay under the quota, it will delete old
versions, layer by layer starting with the biggest version first.
