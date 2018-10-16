.. _getting_started:

Getting started
================

This guide will walk you through the steps of synchronizing your first files
over ``brig``. You will learn about the concepts behind it along the way.
Everything is hands on, so make sure to open a terminal. For now, ``brig`` has
no other user interface.

Precursor: The help system
--------------------------

``brig`` has some built-in helpers to serve as support for your memory. Before
you dive into the actual commands, you should take a look at them.

Built-in documentation
~~~~~~~~~~~~~~~~~~~~~~

Every command offers detailed built-in help, which you can view using the
``brig help`` command:

.. code-block:: bash

    $ brig help stage
    NAME:
       brig stage - Add a local file to the storage

    USAGE:
       brig stage [command options] (<local-path> [<path>]|--stdin <path>)

    CATEGORY:
       WORKING TREE COMMANDS

    DESCRIPTION:
       Read a local file (given by »local-path«) and try to read
       it. This is the conceptual equivalent of »git add«. [...]

    EXAMPLES:

       $ brig stage file.png                         # gets added as /file.png
       $ brig stage file.png /photos/me.png          # gets added as /photos/me.png
       $ cat file.png | brig stage --stdin /file.png # gets added as /file.png

    OPTIONS:
       --stdin, -i  Read data from stdin

Shell autocompletion
~~~~~~~~~~~~~~~~~~~~

If you don't like to remember the exact name of each command, you can use
the provided autocompletion. For this to work you have to insert this
at the end of your ``.bashrc``:

.. code-block:: bash

  source $GOPATH/src/github.com/sahib/brig/autocomplete/bash_autocomplete

Or if you happen to use ``zsh``, append this to your ``.zshrc``:

.. code-block:: bash

  source $GOPATH/src/github.com/sahib/brig/autocomplete/zsh_autocomplete

After starting a new shell you should be able to autocomplete most commands.
Try this for example by typing ``brig remote <tab>``.

Open the online documentation
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

By typing ``brig docs`` you'll get a tab opened in your browser with this
domain loaded.

Reporting bugs
~~~~~~~~~~~~~~~

If you need to report a bug you can use a built-in utility to do that. It will
gather all relevant information, create a report and open a tab with the
*GitHub* issue tracker in a browser for you. Only thing left for you is to fill
out some questions in the report and include anything you think is relevant.

.. code-block:: bash

    $ brig bug

To actually create the issue you sadly need an *GitHub* `account <https://github.com/join>`_.

Creating a repository
---------------------

Let's get started with the actual working commands.

You need a central place where ``brig`` stores files you give it. This place is
called a »repository« or short »repo«. Think of it as a database where all
files and some metadata about them is **copied** to. It is important to keep in mind
that ``brig`` **copies** the file and does not do anything with the original file.

By creating a new repository you also generate your identity, under which your
buddies can later **find** and **authenticate** you.

But enough of the grey theory, let's get started:

.. code-block:: bash

    # Create a place where we store our metadata.
    $ mkdir ~/metadata && cd ~/metadata
    $ brig init --repo . alice@wonderland.lit/rabbithole
    27.12.2017/14:44:39 ⚐ Starting daemon from: /home/sahib/go/bin/brig
    ⚠  39 New passphrase:

    Well done! Please re-type your password now:
    ⚠  39 Retype passphrase:

           _____         /  /\        ___          /  /\
          /  /::\       /  /::\      /  /\        /  /:/_
         /  /:/\:\     /  /:/\:\    /  /:/       /  /:/ /\
        /  /:/~/::\   /  /:/~/:/   /__/::\      /  /:/_/::\
       /__/:/ /:/\:| /__/:/ /:/___ \__\/\:\__  /__/:/__\/\:\
       \  \:\/:/~/:/ \  \:\/:::::/    \  \:\/\ \  \:\ /~~/:/
        \  \::/ /:/   \  \::/~~~~      \__\::/  \  \:\  /:/
         \  \:\/:/     \  \:\          /__/:/    \  \:\/:/
          \  \::/       \  \:\         \__\/      \  \::/
           \__\/         \__\/                     \__\/


         A new file README.md was automatically added.
         Use 'brig cat README.md' to view it & get started.
    $ ls
    config.yml  data  gpg.prv  gpg.pub  logs  metadata
    meta.yml  passwd.locked  remotes.yml

The name you specified after the ``init`` is the name that will be shown
to other users and by which you are searchable in the network.
See :ref:`about_names` for more details on the subject.

Once the ``init`` ran successfully there will be a daemon process running the
background. Every other ``brig`` commands will communicate with it via a local
network socket.

Also note that a lot of files were created in the current directory. This is
all part of the metadata that is being used by the daemon that runs in the
background. Please try not to modify them.

Passwords
~~~~~~~~~

You will be asked to enter a new password. The more secure the password is you
entered, the greener the prompt gets [#]_. This password is used to store
the metadata in an encrypted manner on your filesystem and without further
configuration it needs to be re-entered every time you start the daemon. There
are two ways to prevent that:

1. Use a password helper and tell ``brig`` how to get a password from it by using ``-w / --password-helper`` on the ``init`` command.
   We recommend using `pass <https://www.passwordstore.org/>`_  to do that:

   .. code-block:: bash

       # Generate a password and store it in "pass":
       $ pass generate brig/alice -n 20
       # Tell brig how to get the password out of "pass":
       $ brig init -w "pass brig/alice"
       # Now pass will ask you for the master password with
       # a nice dialog whenever one if its passwords is first used.

2. Do not use a password. You can do this by passing ``-x`` to the ``init`` command.
   This is obviously not recommended.

.. note::

    Using a good password is especially important if you're planning to move
    the repo, i.e. carrying it around you on a usb stick. When the daemon shuts
    down it locks and encrypts all files in the repository (including all
    metadata and keys), so nobodoy is able to access them anymore.

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
``/hello.world``. The name was automatically chosen from looking at the
basename. All files in ``brig`` have their own name, possibly differing from
the content of the file they originally came from. Of course, you can also add
whole directories.

If you want to use a different name, you can simply pass the new name as second
argument to ``stage``:

.. code-block:: bash

    $ brig stage /tmp/hello.world /hallo.welt

You also previously saw ``brig cat`` which can be used to get the content of
a file again. ``brig ls`` in contrast shows you a list of currently existing
files, including their size, last modification time, path and pin state [#]_.

Coreutils
---------

You probably already noticed that a lot of commands you'd type in a terminal have
a sibling as ``brig`` command. Here is a short overview of the available commands:

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
    $ brig rm photos

Please refer to ``brig help <command>`` for more information about those. Often
they work a little bit different [#]_ and a bit less surprising than their
counterparts. Also note that there is no ``brig cd`` currently. All paths must
be absolute.

Mounting repositories
---------------------

Of course, using those specialized ``brig`` commands all day can be annoying
and feels not very seamless, especially when being used to tools like file
browsers. Indeed, those commands are only supposed to serve as a low-level way
of interacting with ``brig`` and as means for scripting own workflows.

For your daily workflow it is far easier to mount all files known to ``brig``
to a directory of your choice and use it with your normal tools. To accomplish
that ``brig`` supports a FUSE filesystem that can be controlled via the
``mount`` and ``fstab`` commands. Let's look at ``brig mount``:

.. code-block:: bash

   $ mkdir ~/data && cd ~/data
   $ brig mount ~/data
   $ cat hello-world
   Hello World
   $ echo 'Salut le monde!' > salut.monde
   # There is no difference between brig's "virtual view"
   # and the conents of the mount:
   $ brig cat salut.monde
   Salut le monde!

You can use this directory exactly [#]_ like a normal one. You can have any
number of mounts. This proves especially useful when only mounting
a subdirectory of ``brig`` (let's say ``Public``) with the ``--root`` option of
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

It can get a little annoying of course when having to manage all mounts
yourself. It would be nice to have some *typical* mounts you'd like to have
always and it should be only one command to mount or unmount all of them, kind
of what ``mount -a`` does. That's what ``brig fstab`` is for:

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

*Caveats:* The FUSE filesystem is not (yet) perfect. Keep those points in mind:

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

Remotes
-------

Until now, all our operations were tied only to our local computer. But
``brig`` is a synchronization tool and that would be hardly very useful without
supporting other peers.

Every peer possesses two things that identifies him:

- **A human readable name:** This name can be choose by the user and can take
  pretty much any form, but we recommend to sticking for a form that resembles
  an extended email [#]_ like »ali@woods.org/desktop«. The name is **not**
  unique! In theory everyone could take it and it is therefore only used for
  display purposes.
- **A unique fingerprint:** This serves both as address for a certain repository and as certificate of identity.
  It is long and hard to remember, which is the reason why ``brig`` offers to loosely link a human readable to it.

If we want to find out what our name and fingerprint is, we can use the ``brig
whoami`` command to ask existential questions:

.. code-block:: bash

    # NOTE: The hash will look different for you:
    $ brig whoami
    ali@home.cz/desktop QmTTJbkfG267gidFKfDTV4j1c843z4tkUG93Hw8r6kZ17a:SEfXUDvKzjRPb4rbbkKqwfcs1eLkMwUpw4C35TJ9mdtWnUHJaeKQYxjFnu7nzrWgU3XXHoW6AjvBv5FcwyJjSMHu4VR4f

.. note::

    The fingerprint consists of two hashes divided by a colon (:). The first
    part is the identity of your ``ipfs`` node, the second part is the
    fingerprint of a keypair that was generated by ``brig`` during init and
    will be used to authenticate other peers.

When we want to synchronize with another repository, we need to exchange fingerprints.
There are three typical scenarios here:

- Both repositories are controlled by you. In this case you can simple execute
  ``brig whoami`` on both repositories.
- You want to sync with somebody you know well. In this case you should both
  execute ``brig whoami`` and send it over a trusted sidechannel. Personally,
  I use a `secure messenger like Signal <https://signal.org>`_, but you can
  also use any channel you like, including encrypted mail or meeting up with
  the person in question.
- You don't know each other. Get to know each other and the proceed like in the
  second point.

.. todo::

    Mention ``brig net locate``

Once you have exchanged the fingerprints, you add each other as **remotes**.
Let's call the other side *bob*: [#]_

.. code-block:: bash

	$ brig remote add bob \
		QmUDSXt27LbCCG7NfNXfnwUkqwCig8RzV1wzB9ekdXaag7:
		SEfXUDSXt27LbCCG7NfNXfnwUkqwCig8RzV1wzB9ekdXaag7wEghtP787DUvDMyYucLGugHMZMnRZBAa4qQFLugyoDhEW

*Bob* has do the same on his side. Otherwise the connection won't be
established, because the other side won't be authenticated. Adding somebody as
remote is the way to authenticate them.

.. code-block:: bash

	$ brig remote add ali \
        QmTTJbkfG267gidFKfDTV4j1c843z4tkUG93Hw8r6kZ17a:
        SEfXUDvKzjRPb4rbbkKqwfcs1eLkMwUpw4C35TJ9mdtWnUHJaeKQYxjFnu7nzrWgU3XXHoW6AjvBv5FcwyJjSMHu4VR4f

Thanks to the fingerprint, ``brig`` now knows how to reach the other repository over the network.

.. todo::

    Rework this whole section:

    TODO: Network intermezzo?

    TODO: Auto accept?

The remote list can tell us if a remote is online:

.. code-block:: bash

    $ brig remote list
    NAME   FINGERPRINT  ROUNDTRIP  LASTSEEN
    bob    QmUDSXt27    0s         ✔ Apr 16 17:31:01
    $ brig remote ping bob
    ping to bob: ✔ (0.00250ms)

Nice. Now we know that bob is online and also that he authenticated us.
Otherwise ``brig remote ping bob`` would have failed. (TODO: This needs some cleanup)

.. note:: About open ports:

   While ``ipfs`` tries to do it's best to avoid having the user to open ports
   in his firewall/router. This mechanism might not be perfect though and maybe
   never is. If any of the following network operations might not work it might
   be necessary to open the ports 4001 - 4005 and/or enable UPnP. For security
   reasons we recommend to only open the required ports explicitly and not to
   use UPnP unless necessary though. This is only necessary if the computers
   you're using ``brig`` on are not in the same network anyways.

.. _about_names:

Choosing and finding names
~~~~~~~~~~~~~~~~~~~~~~~~~~

You might wonder what the name you pass to ``init`` is actually for. As
previously noted, there is no real restriction for choosing a name, so all of
the following are indeed valid names:

- ``ali``
- ``ali@woods.org``
- ``ali@woods.org/desktop``
- ``ali/desktop``

It's however recommended to choose a name that is formatted like
a XMPP/Jabber-ID. Those IDs can look like plain emails, but can optionally have
a »resource« part as suffix (separated by a »/« like ``desktop``). Choosing
such a name has two advantages:

- Other peers can find you by only specifying parts of your name.
  Imagine all of the *Smith* family members use ``brig``, then they'd possibly those names:

  * ``dad@smith.org/desktop``
  * ``mom@smith.org/tablet``
  * ``son@smith.org/laptop``

  When ``dad`` now sets up ``brig`` on his server, he can use ``brig net locate
  -m domain 'smith.org'`` to get all fingerprints of all family members. Note
  however that ``brig net locate`` **is not secure**. Its purpose is solely
  discovery, but is not able to verify that the fingerprints really correspond
  to the persons they claim to be. This due to the distributed nature of
  ``brig`` where there is no central or federated authority that coordinate
  user name registrations. So it is perfectly possible that one name can be
  taken by several repositories - only the fingerprint is unique.

  .. todo::

    Provide output of the locate command and verify this scenario works fine.

- Later development of ``brig`` might interpret the user name and domain as
  email and might use your email account for verification purposes.

Having a resource part is optional, but can help if you have several instances
of ``brig`` on your machines. i.e. one username could be
``dad@smith.org/desktop`` and the other ``dad@smith.org/server``.

Syncing
-------

Before we move on to do our first synchronization, let's recap what we have don so far:

- Create a repository (``brig init <name>``) - This needs to be done only once.
- Create optional mount points (``brig fstab add <name> <path>``) - This needs to be done only once.
- Find & add remotes (``brig remote add``) - This needs to be done once for each peer.
- Add some files (``brig stage <path>``) - Do as often as you like.

As you see, there is some initial setup work, but the actual syncing is pretty
effortless now. Before we attempt to sync with anybody, it's always a good idea
to see what changes they have. We can check this with ``brig diff <remote>``:

.. code-block:: bash

    # The "--missing" switch also tells us what files they don't have:
    $ brig diff bob --missing
    •
    ├── _ hello.world
    ├── + videos/
    └── README.md ⇄ README.md

This output resembles the one we saw from ``brig tree`` earlier.
Each node in this tree tells us about something that would
happen when we merge. The prefix of each file and the color in the terminal
indicate what would happen with this file. Refer to the table below to see what
prefix relates to what action:

====== ====================================================================
Symbol Description
====== ====================================================================
``+``  The file is only present on the remote side.
``-``  The file was removed on the remote side.
``→``  The file was moved to a new location.
``*``  This file was ignored because we chose to, due to our settings.
``⇄``  Both sides have changes, but they are compatible and can be merged.
``⚡``  Both sides have changes, but they are incompatible and result in conflicts.
``_``  The file is missing on the remote side (output needs to be enabled with ``--missing``)
====== ====================================================================

.. note::

    ``brig`` does not do any actual diffs between files. It does not care a lot about the content.
    It only records how the file metadata changes and what content the file has at a certain point.

If you prefer a more traditional view, similar to ``git``, you can use
``--list`` on ``brig diff``.

So in the above output we can tell that *Bob* added the directory
``/videos``, but does not possess the ``/hello.world`` file. He also
apparently modified ``README.md``, but since we did not, it's safe for us to
take over his changes. If we sync now we will get this directory from him:

.. code-block:: bash

    $ brig sync bob
    $ brig ls
    SIZE   MODTIME          OWNER    PATH                      PIN
    443 B  Dec 27 14:44:44  sahib    /README.md                🖈
    443 B  Dec 27 14:44:44  bob      /README.md.conflict.0
    12 B   Dec 27 15:14:16  sahib    /hello.world              🖈
    32 GB  Dec 27 15:14:16  bob      /videos                   🖈

You might notice that the ``sync`` step took only around one second, even
though ``/videos`` is 32 GB in size. This is because ``sync`` *does not
transfer actual data*. It only transferred the metadata, while the actual data
will only be loaded when required. This sounds a little inconvenient at first.
When I want to watch the video, I'd prefer to have it cached locally before
viewing it to avoid stuttering playback. If you plan to use that, you're free
to do so using pinning (see :ref:`pinning-section`)

Data retrieval
~~~~~~~~~~~~~~

If the data is not on your local machine, where is it then? Thanks to ``ipfs``
it can be transferred from any other peer that caches this particular content.
Content is usually cached when the peer either really stores this file or if
this peer recently used this content. In the latter case it will still be
available in its cache. This property is particularly useful when having
a small device for viewing data (e.g. a smartphone) and a big machine that acts
as storage server (e.g. a desktop).

How are the files secure then if they essentially could be everywhere? Every
file is encrypted by ``brig`` before giving it to ``ipfs``. The encryption key
is part of the metadata and is only available to the peers that you chose to
synchronize with. Think of each brig repository only as a cache for the whole
network it is in.

Partial synchronisation
~~~~~~~~~~~~~~~~~~~~~~~

Sometimes you only want to share certain things with certain people. You
probably want to share all your ``/photos`` directory with your significant
other, but not with your fellow students where you maybe want to share the
``/lectures`` folder. In ``brig`` you can define what folder you want to share
with what remote. If you do not limit this, all folders will be open to
a remote by default.


.. todo::

    * Write docs for brig remote folders.
    * Finish brig remote folder handling on the command line.

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

.. todo::

    * Explain the implications of pinning when syncing files and other operations like reset.
    * Add a command example and an example that shows what "brig gc" does.

If you never pin something explicitly, only the newest version of all files
will be stored locally. If you decide that you need older versions, you can pin
them explictly, so brig cannot unpin them implicitly. For this you should also
look into the ``brig pin set`` and ``brig pin clear`` commands, which are
similar to ``brig pin add`` and ``brig pin rm`` but can operate on whole commit
ranges.

Garbage collection
~~~~~~~~~~~~~~~~~~

Strongly related to pinning is garbage collection. This is normally being run
for you every few minutes, but you can also trigger it manually via the ``brig
gc`` command. While not usually needed, it can help you understand how ``brig``
works internally as it shows what hashes it throws away.

Version control
---------------

One key feature of ``brig`` over other synchronisation tools is the built-in
and quite capable version control. If you already know ``git`` that's a plus
for this chapter since a lot of stuff will feel similar. This is not surprise,
since ``brig`` implements something like ``git`` internally. Don't worry,
knowing ``git`` is however not needed at all for this chapter.

Key concepts
~~~~~~~~~~~~

I'd like you to keep the following mantra in your head when thinking
about versioning (repeating before you go to sleep may or may not help):

**Metadata and actual data are separated.** This means that a repository may
contain metadata about many files, including older versions of them. However,
it is not guaranteed that a repository caches all actual data for each file or
version. This is solely controlled by pinning described in the section before.
If you check out earlier versions of a file, you're always able to see the
metadata of it, but being able to view the actual data depends on having a peer
that is being able to deliver the data in your network (which might be
yourself). So in short: ``brig`` **only versions metadata and links to the
respective data for each version**.

This is a somewhat novel approach to versioning, so feel free to re-read the
last paragraph, since I've found that it does not quite fit what most people
are used to. Together with pinning this offers a high degree of freedom on
how you can decide what repositories store what data.

You can invoke ``brig info`` to see what metadata is being saved per file version:

.. code-block:: bash

    $ brig show README.md
    Path          /README.md
    User          ali
    Type          file
    Size          832 bytes
    Inode         4
    Pinned        yes
    Explicit      no
    ModTime       2018-10-14T22:46:00+02:00
    Tree Hash     SEfXUE2YhBFALY7EQd1BbYFugugqipCeKmadx7wMo5SRdNjNZhaCV9W77vs8aYjvTnB8uvC4ZKi5znaq9iGaKZyTyjZv6
    Content Hash  SEfXUDMbsF97A5vgf52aXsdVEVhGPKFC2QUU3946yoFTL3EsqjRJHTXNZSz1vhKegrmwBKQFghvREQoNUVRv7Hx6b8a1M
    Backend Hash  QmPvNjR1h56EFK1Sfb7vr7tFJ57A4JDJS9zwn7PeNbHCsK


Most of it should be no big surprise. It might be a small surprise that three
hashes are stored per file. The ``Backend Hash`` is really the link to the
actual data. If you'd type ``ipfs cat
QmPvNjR1h56EFK1Sfb7vr7tFJ57A4JDJS9zwn7PeNbHCsK`` you might get the encrypted
version of your file dumped to your terminal. The ``Content Hash`` is being
calculated before the encryption and is the same for two files with the same
content. The ``Tree Hash`` is a hash that uniquely identifies this specific
node. The ``Inode`` is unique to a file and is also used in the FUSE
filesystem.

Commits
~~~~~~~

Now that we know that only metadata is versioned, we have to ask »what is the
smallest unit of modification that can be saved?« This smallest unit is
a commit. A commit can be seen as a snapshot of a repository.

The command ``brig log`` shows you a list of commits that were made already:

.. code-block:: bash

          -      Sun Oct 14 22:46:00 CEST 2018 • (curr)
    SEfXUDozvTHH Sun Oct 14 22:46:00 CEST 2018 user: Added ali-file (head)
    SEfXUASkpNy4 Sun Oct 14 22:46:00 CEST 2018 user: Added initial README.md
    SEfXUEru1pLi Sun Oct 14 22:46:00 CEST 2018 initial commit (init)


Each commit is identified by a hash (e.g. ``SEfXUDozvTHH``) and records the
time when it was created. Apart from that, there is a message that describes
the commit in some way. In contrast to ``git``, **commits are rarely done by
the user themselve**. More often they are done by ``brig`` when synchronizing.

All commits form a long chain (**no branches**, just a linear chain) with the
very first empty commit called ``init`` and the still unfinished commit called
``curr``. Directly below ``curr`` there is the last finished commit called ``head``.

.. note::

    ``curr`` is what ``git`` users would call the staging area. While the staging area
    in ``git`` is "special", the ``curr`` commit can be used like any other one, with
    the sole difference that it does not have a proper hash yet.

Sometimes you might want to do a snapshot or »savepoint« yourself. In this case
you can do a commit yourself:

.. code-block:: bash

    $ brig touch A_NEW_FILE
    $ brig commit -m 'better leave some breadcrumbs'
    $ brig log | head -n 2
          -      Mon Oct 15 00:27:37 CEST 2018 • (curr)
    SEfXUDkdjUND Sun Oct 14 22:46:00 CEST 2018 user: better leave some bread crumbs (head)

This snapshot can be useful later if you decide to revert to a certain version.
The hash of the commit is of course hard to remember, so if you need it very often, you can
give it a tag yourself. Tags are similar to the names, ``curr``, ``head`` and ``init`` but
won't be changed by ``brig`` and won't move therefore:

.. code-block:: bash

    # instead of "SEfXUDkdjUND" you also could use "head" here:
    $ brig tag SEfXUDkdjUND breadcrumbs
    $ brig log | grep breadcrumps
    $ SEfXUDkdjUND Sun Oct 14 22:46:00 CEST 2018 user: better leave some bread crumbs (breadcrumbs, head)


File history
~~~~~~~~~~~~

Each file and directory in ``brig`` maintains its own history. Each entry of
this history relates to exactly one distinct commit. In the life of a file or
directory there are four things that can happen to it:

- *added:* The file was added in this commit.
- *moved:* The file was moved in this commit.
- *removed:* The file was removed in this commit.
- *modified:* The file's content (i.e. hash changed) was altered in this commit.

You can check an individual file or directorie's history by using the ``brig history`` command:

.. code-block:: bash

    # or "hst" for short:
    $ brig hst README.md
    CHANGE  FROM  TO              WHEN
    added   INIT  SEfXUASkpNy4    Oct 14 22:46:00
    $ brig mv README.md README_LATER.md
    $ brig hst README_LATER.md
    CHANGE  FROM  TO            HOW                           WHEN
    moved   HEAD  CURR          /README.md → /README_LATER.md Oct 15 00:27:37
    added   INIT  SEfXUASkpNy4                                Oct 14 22:46:0

As you can see, you will be shown one line per history entry. Each entry
denotes which commit the change was in. Some commits were nothing was changed
will be jumped over except if you pass ``--empty``.

Viewing differences
~~~~~~~~~~~~~~~~~~~

If you're interested what changed in a range of commits, you can use the ``brig
diff`` command as shown previously. The ``-s`` (``--self``) switch says that it
should only look at own commits and not compare any remotes.

.. code-block:: bash

    # Let's compare the commit hashes from above:
    $ brig diff -s SEfXUDkdjUND SEfXUDozvTHH
    •
    └── + A_NEW_FILE

Often, those hashes are quite hard to remember and annoying to look up. That's
why you can the special syntax ``<tag or hash>^`` to denote that you want to go
»one commit up«:

.. code-block:: bash

    brig diff -s head head^
    •
    └── + A_NEW_FILE
    # You can also use this several times:
    brig diff -s head^^^ head^^^^^
    •
    └── + README.md

If you just want to see what you changed since ``head``, you can simply type ``brig diff``.
This is the same as ``brig diff -s curr head``:

.. code-block:: bash

    $ brig diff
    •
    └── README.md → README_LATER.md
    $ brig diff -s curr head
    •
    └── README.md → README_LATER.md


Reverting to previous state
~~~~~~~~~~~~~~~~~~~~~~~~~~~

Until now we were only looking at the version history and didn't modify it. The
most versatile command to do that is ``brig reset``. It is able to revert
changes previously made:

.. code-block:: bash

    # Reset to the "init" commit (the very first and empty commit)
    $ brig reset init
    $ brig ls  # nothing, it's empty.


The key here is that you did not loose any history:

.. code-block:: bash

    $ brig log | head -2
           -     Mon Oct 15 00:51:12 CEST 2018 • (curr)
    SEfXUDkdjUND Sun Oct 14 22:46:00 CEST 2018 user: better leave some bread crumbs (breadcrumbs)


As you can see, we still have the previous commits. ``brig revert`` did not
thing more than restoring the state of ``init`` and put that result in
``curr``. This also means that you can't really *modify* history. But you can
revert it. Let's revert your complete wipe-out:

.. code-block:: bash

    $ brig reset breadcrumbs


Now everything is as we left it. ``brig reset`` cannot only restore old
commits, but individual files and directories:


.. code-block:: bash

    $ brig reset head^^

.. note::

    It is a good idea to do a ``brig commit`` before a ``brig reset``. Since it
    modifies ``curr`` you might loose uncommitted changes. It will warn you
    about that, but you can overwrite that warning with ``--force``. If you did
    a ``brig commit`` you can simply use ``brig reset head`` to go back to the
    last good state.


Other commands
~~~~~~~~~~~~~~

There are a few other commands, but they are not (yet) very useful for most end
users. Therefore they will not be explained in depth to save you some mental
space. The commands in question are:

- ``brig become``: View the metadata of another remote. Good for debugging.
- ``brig daemon``: Start the daemon manually. Good for init systems like ``systemd``.
- ``brig net``: Commands to modify the network status and find other peers.
- ``brig edit``: Edit a file in brig with the ``$EDITOR``.
- ``brig fetch``: Manually trigger the fetching of a remote's metadata.
- ``brig tar``: Output files and directories as tar archive. Useful to output whole directories.

Please use ``brig help <command>`` to find out more about them if you're interested.

Configuration
-------------

Quite a few details can be configured in a different way to your liking. ``brig
config`` is the command that allows you to list, get and set individual
configuration values. Each config entry already brings some documentation that
tells you about its purpose:

.. code-block:: bash

    $ brig config ls
    [... output truncated ...]
    fs.sync.ignore_moved: false (default)
    Default:       false
    Documentation: Do not move what the remote moved
    Needs restart: no
    [... output truncated ...]
    $ brig config get repo.password_command
    pass brig/repo/password
    $ brig config set repo.password_command "pass brig/repo/my-password"

Running the daemon and viewing logs
-----------------------------------

As discussed before, the daemon is being started on demand in the background.
Subsequent commands will then use the daemon. For debugging purposes it can be useful
to run in the daemon in the foreground. You can do this with the ``brig daemon`` commands:

.. code-block:: bash

    # Make sure no prior daemon is running:
    $ brig daemon quit
    # Start the daemon in the foreground and log to stdout:
    $ brig daemon launch -s

The last step will ask for your password if you did not set a password helper
program. If you want to quit the instance, either just hit CTRL-C or type
``brig daemon quit`` into another terminal window.

Logging
~~~~~~~

Unless you pass the ``-s`` (``--log-to-stdout`` flag) all logs are being piped
to the system log. You can follow the log like this:

.. code-block:: bash

    # The actual daemon log:
    $ journalctl -ft brig

    # The ipfs log:
    $ journalctl -ft brig-ipfs

This assumes you're using a ``systemd``-based distribution. If not, refer to
the documentation of your syslog daemon.

Using several repositories in parallel
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

It can be useful to run more than one instance of the ``brig`` daemon in
parallel. Either for testing purposes or as actual production configuration. If
you're planning to do that it is advisable to be always explicit about the port
number you're using. Here's an example how you can run two daemons at the same
time:

.. code-block:: bash

    # It might be a good idea to keep that in your .bashrc:
    alias brig-ali='brig --port 6666'
    alias brig-bob='brig --port 6667'

    # Subsitute your password helper here:
    brig-ali --repo /tmp/ali init ali -w "echo brig/repo/ali"
    brig-bob --repo /tmp/bob init bob -w "echo brig/repo/bob"

    # Now you can use them normally,
    # e.g. by adding them as remotes each:
    brig-ali remote add bob $(brig-bob whoami -f)
    brig-bob remote add ali $(brig-ali whoami -f)

-------

.. [#] This uses `Dropbox's password strength library »zxcvbn« <https://github.com/dropbox/zxcvbn>`_.


.. [#] Pinning and pin states are explained :ref:`pinning-section` and are not important for now.

.. [#] ``brig rm`` for example deletes directories without needing a ``-r`` switch.

.. [#] Well almost. See the *Caveats* below.

.. [#] To be more exact, it resembles an `XMPP or Jabber-ID <https://en.wikipedia.org/wiki/Jabber_ID>`_.

.. [#] The name you choose as remote can be anything you like and does not need
       to match the name the other person chose for themselves. It's not a bad
       idea though.

