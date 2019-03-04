Mounting repositories
---------------------

Using those specialized ``brig`` commands might not feel very seamless,
especially when being used to tools like file browsers. And indeed, those
commands are only supposed to serve as a low-level way of interacting with
``brig`` and as means for scripting own workflows.

For your daily workflow it is far easier to mount all files known to ``brig``
to a directory of your choice and use it with your normal tools. To accomplish
that ``brig`` supports a FUSE filesystem that can be controlled via the
``mount`` and ``fstab`` commands. Let's look at ``brig mount``:

.. code-block:: bash

   $ mkdir ~/data
   $ brig mount ~/data
   $ cd ~/data
   $ cat hello-world
   Hello World
   $ echo 'Salut le monde!' > salut-monde.txt
   # There is no difference between brig's "virtual view"
   # and the contents of the mount:
   $ brig cat salut-monde.txt
   Salut le monde!

You can use this directory like a normal one (except for the CAVEATS below). You can have any
number of mounts. This proves especially useful when only mounting a subdirectory of ``brig``
(let's say ``Public``) with the ``--root`` option of
``brig mount`` and mounting all other files as read only (``--readonly``).

.. code-block:: bash

    $ brig mount ~/data --readonly
    $ brig mkdir /writable
    $ brig touch /writable/please-edit-me
    $ mkdir ~/rw-data
    $ brig mount ~/rw-data --root /writable
    $ echo 'writable?' > ~/data/test
    read-only file system: ~/data/test
    $ echo 'writable!' > ~/rw-data/test
    $ cat ~/rw-data/test
    writable!

An existing mount can be removed again with ``brig unmount <path>``:

.. code-block:: bash

    $ brig unmount ~/data
    $ brig unmount ~/rw-data
    $ brig rm writable

.. _permanent-mounts:

Making mounts permanent
~~~~~~~~~~~~~~~~~~~~~~~

It can get a little annoying when having to manage several mounts yourself. It
would be nice to have some *typical* mounts you'd like to have always and it
should be only one command to mount or unmount all of them, kind of what
``mount -a`` does. That's what ``brig fstab`` is for:

.. code-block:: bash

    $ brig fstab add tmp_rw_mount /tmp/rw-mount
    $ brig fstab add tmp_ro_mount /tmp/ro-mount -r
    $ brig fstab
    NAME          PATH           READ_ONLY  ROOT  ACTIVE
    tmp_ro_mount  /tmp/ro-mount  yes        /
    tmp_rw_mount  /tmp/rw-mount  no         /
    $ brig fstab apply
    $ brig fstab
    NAME          PATH           READ_ONLY  ROOT  ACTIVE
    tmp_ro_mount  /tmp/ro-mount  yes        /     ✔
    tmp_rw_mount  /tmp/rw-mount  no         /     ✔
    $ brig fstab apply -u
    NAME          PATH           READ_ONLY  ROOT  ACTIVE
    tmp_ro_mount  /tmp/ro-mount  yes        /
    tmp_rw_mount  /tmp/rw-mount  no         /

Et Voilà, all mounts will be created and mounted once you enter ``brig fstab
apply``. The opposite can be achieved by executing ``brig fstab apply --unmount``.
On every restart of the daemon, all mounts are mounted by default, so the only
thing you need to make sure is that the daemon is running.

*CAVEATS:* The FUSE filesystem is not yet perfect. Keep those points in mind:

- **Performance:** Writing to FUSE is currently somewhat *memory and CPU
  intensive*. Generally, reading should be fast enough for most basic use
  cases, but also is not enough for high performance needs. If you need to edit
  a file many times, it is recommended to copy the file somewhere to your local
  storage (e.g. ``brig cat the_file > /tmp/the_file``), edit it there and save
  it back for syncing purpose. Future releases will work on optimizing the
  performance.
- **Timeouts:** Although it tries not to look like one, we're operating on
  a networking filesystem. Every file you access might come from a different
  computer. If no other machine can serve this file we might block for a long
  time, causing application hangs and general slowness. This is a problem that
  still needs a proper solution and leaves much to be desired in the current
  implementation.
