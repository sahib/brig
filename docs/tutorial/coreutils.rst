Adding & Viewing files
----------------------

Now let's add some files to ``brig``. We do this by using ``brig stage``. It's
called ``stage`` because all files first get added to a staging area. If you
want, and are able to remember that easier, you can also use ``brig add``.

.. code-block:: bash

    $ echo "Hello World" > /tmp/hello.world
    $ brig stage /tmp/hello.world
    $ brig cat hello.world
    Hello World
    $ brig ls
    SIZE   MODTIME          PATH          PIN
    443 B  Dec 27 14:44:44  /README.md     🖈
    12 B   Dec 27 15:14:16  /hello.world   🖈

This adds the content of ``/tmp/hello.world`` to a new file in ``brig`` called
``/hello.world``. The name was automatically chosen from looking at the base
name of the added file. All files in ``brig`` have their own name, possibly
differing from the content of the file they originally came from. Of course,
you can also add whole directories.

If you want to use a different name, you can simply pass the new name as second
argument to ``stage``:

.. code-block:: bash

    $ brig stage /tmp/hello.world /hallo.welt

You also previously saw ``brig cat`` which can be used to get the content of
a file again. ``brig ls`` in contrast shows you a list of currently existing
files, including their size, last modification time, path and pin state [#]_.

One useful feature of ``brig cat`` is that you can output directories as well.
When specifying a directory as path, a ``.tar`` archive is being outputted.
You can use that easily to store whole directories on your disk or archive
in order to send it to some client for example:

.. code-block:: bash

   # Create a tar from root and unpack it to the current directory.
   $ brig cat | tar xfv -
   # Create .tar.gz out of of the /photos directory.
   $ brig cat photos | gzip -f > photos.tar.gz

.. [#] Pinning and pin states are explained :ref:`pinning-section` and are not important for now.

Coreutils
---------

You probably already noticed that a lot of commands you'd type in a terminal on
a normal day have a sibling as ``brig`` command. Here is a short overview of
the available commands:

.. code-block:: bash

    $ brig mkdir photos
    $ brig touch photos/me.png
    $ brig tree
    • 🖈
    ├──photos 🖈
    │  └── me.png 🖈
    ├── README.md 🖈
    └── hello.world 🖈

    2 directories, 2 files
    $ brig cp photos/me.png photos/moi.png
    $ brig mv photos/me.png photos/ich.png
    # NOTE: There is no "-r" switch. Directories are always deleted recursively.
    $ brig rm photos

Please refer to ``brig help <command>`` for more information about those.
Sometimes they work a little bit different [#]_ and a bit less surprising than
their counterparts. Also note that there is no ``brig cd`` currently. All paths
must be absolute.

.. [#] ``brig rm`` for example deletes directories without needing a ``-r`` switch.
