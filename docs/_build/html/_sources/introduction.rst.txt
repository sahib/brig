.. _getting_started:

Getting started
================

This guide will walk you through the steps of synchronizing your first files
over ``brig``. It's hand's on, so make sure to open a terminal.
We'll explain all import concepts along the way.

Precursor: The help system
--------------------------

``brig`` has some built-in commands to help you.
Before you dive into the actual commands, you should take a look at them:


Open the online documentation
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

This opens a new browser window with this site.
URLs are hard to remember, so we got you covered:

.. code-block:: bash

    $ brig docs

Built-in documentation
~~~~~~~~~~~~~~~~~~~~~~

Every command offers detailled built-in help,
which you can view using the ``brig help`` command:

.. code-block:: bash


    $ brig help remote
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

       $ brig stage file.png                   # gets added as /file.png
       $ brig stage file.png /photos/me.png    # gets added as /photos/me.png
       $ cat file.png | brig --stdin /file.png # gets added as /file.png

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

After starting a new shell you should be able to do this:

.. code-block:: bash

    # There's builtin autocompletion if you source
    # autocomplete/bash_autocomplete
    $ brig remote <tab>
    add     clear   edit    list    ping    remove

Typo suggestion
~~~~~~~~~~~~~~~

This is a rather silly, little feature but if you mistype a command, you get
suggestion on what you likely meant to type:

.. code-block:: bash

    $ brig remot
    `remot` is not a valid command.

    Did you maybe mean one of those?
      * reset
      * mount
      * rm
      * remote

Reporting bugs
~~~~~~~~~~~~~~~

If you need to report a bug you can use a built-in utility to do that. It will
gather all relevant information, create a report and open a tab with the
*GitHub* issue tracker in a browser for you. Only thing left for you is to fill
out some questions in the report (and possibly create a *GitHub* account
first):

.. code-block:: bash

    $ brig bug

Creating a repository
---------------------

You need a central place where your files are stored and ``brig`` calls this
place the *repository*. Note that this is not directly comparable to what other
tools calls the *Sync folder*. Rather think of it as the ``.git`` folder of
a ``git``-repository: A place where all internal state, data and metadata of
``brig`` is stored.

By creating a new repository you also generate your identity, under which
your buddies can later find *and* authenticate you.

But enough of the grey theory, let's get started:

.. code-block:: bash

    $ mkdir ~/metadata && cd ~/metadata
    $ brig init alice@wonderland.lit/rabbithole
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

You will be asked to enter a new password. *»Why«* you ask? This password is
used to store your data in an encrypted manner on your harddisk. This is
especially important if you think about creating the repository on a portable
media (e.g. usb sticks). If you still choose to disable this security feature
you're free to do so by passing ``-x`` directly before the ``init`` subcommand.
When typing the password, you will notice that the prompt changes color. The
more secure (measured by Dropbox's ``zxcvbn`` library) the password is, the
closer to green the prompt gets.

Also note that a lot of files were created in the current directory.
This is all part of the metadata that is being used by the daemon that runs
in the background.

Adding & Viewing files
----------------------

Phew, that was a lot of text, but there was not any real action yet.
Let's change that by adding some files to ``brig``:

.. code-block:: bash

    $ echo "Hello World" > /tmp/hello.world
    $ brig stage /tmp/hello.world
    $ brig cat hello.world
    Hello World
    $ brig ls
    SIZE   MODTIME          PATH          PIN
    443 B  Dec 27 14:44:44  /README.md     🖈
    12 B   Dec 27 15:14:16  /hello.world   🖈

You might have noticed that the »hello.world« file was stored in ``brig`` without the
full path (»/tmp/hello.world«). This is done on purpose, since you should imagine all
added files live under an own root. You can however give the file a new name while adding it:

.. code-block:: bash

    $ brig stage /tmp/hello.world /hallo.welt

Mounting repositories
---------------------

There are subcommands that act very similar to ``mkdir``, ``rm`` and ``mv``.
While those surely are useful, it's not a very native feel of handling files.
That's why you can mount all files kown to ``brig`` to a special folder:

.. code-block:: bash

   $ mkdir ~/data && cd ~/data
   $ brig mount ~/data
   $ cat hello-world
   Hello World


You can use this directory (almost) exactly like a normal one.
We recommend though, that you shouldn't do any heavy editing inside of the folder
and use it more like a »transfer box« for efficiency reasons.

Remotes
-------

Until now, all files where only local. How do we even talk to other peers? This
is done by adding them as »remote«. Every repository you are using has
a user-chosen name (»alice@wonderland.lit/rabbithole«) and a unique
fingerprint that was generated during ``init``. Let's see what our own fingerprint is:


.. code-block:: bash

    # The hash will most likely look different for you:
    $ brig whoami
    alice@wonderland.lit/rabbithole QmTTJbkfG267gidFKfDTV4j1c843z4tkUG93Hw8r6kZ17a:SEfXUDvKzjRPb4rbbkKqwfcs1eLkMwUpw4C35TJ9mdtWnUHJaeKQYxjFnu7nzrWgU3XXHoW6AjvBv5FcwyJjSMHu4VR4f

The fingerprint consists of two hashes divided by a colon (:). The first part
is the identity of your ``ipfs`` node, the second part is the fingerprint of
a keypair that was generated by ``brig`` and will be used to authenticate other
peers.

Now let's assume another user (let's call him Bob) wants to synchronize files
with Alice. Both sides now need to share the information printed by ``brig
whoami`` over a secure side channel. This side channel could be one of the
following:

- Encrpyted mail.
- A secure instant messenger of your choice.
- Any *insecure* channel, as long you call or meet the person later and you
  validate at least a few digits of his fingerprint.

Once you have exchanged the fingerprints, *Alice* can add *Bob*:

.. code-block:: bash

	$ brig remote add bob \
		QmUDSXt27LbCCG7NfNXfnwUkqwCig8RzV1wzB9ekdXaag7:
		SEfXUDSXt27LbCCG7NfNXfnwUkqwCig8RzV1wzB9ekdXaag7wEghtP787DUvDMyYucLGugHMZMnRZBAa4qQFLugyoDhEW


*Bob* can do the same on his side:

.. code-block:: bash

	$ brig remote add alice \
        QmTTJbkfG267gidFKfDTV4j1c843z4tkUG93Hw8r6kZ17a:
        SEfXUDvKzjRPb4rbbkKqwfcs1eLkMwUpw4C35TJ9mdtWnUHJaeKQYxjFnu7nzrWgU3XXHoW6AjvBv5FcwyJjSMHu4VR4f

After doing so ``brig`` can figure out the rest (i.e. how to actually reach the
node over the network itself). Remember that this mechanism might seem
inconvinient at first, but it's the only way for you to actually check if
someone is truly the person he claims to be.

If both sides are up & running, we can check if we can reach the other side:

.. code-block:: bash

	$ brig remote list
    NAME   FINGERPRINT  ROUNDTRIP  LASTSEEN
    bob    QmUDSXt27    ∞          ✔ Apr 16 17:31:01
	# Yep that works.
	$ brig remote ping bob
    ping to bob: ✔ (0.00250ms)

Cool, we are ready to reach them. Note that ``brig remote list`` only shows if
a another node is really online. ``brig remote ping <name>`` sends an actual
message to them which will only be replied back if they mutually authenticated
us!

.. note:: About open ports:

   While ``ipfs`` tries to do it's best to avoid having the user to open ports
   in his firewall/router. This mechanism might not be perfect though (and
   maybe never is). If any of the following network operations might not work
   it might be necessary to open the ports 4001 - 4005 or enable UPnP. For
   security reasons we recommend to only open the required ports explicitly and
   not to use UPnP. This only is necessary if the computers you're using
   ``brig`` on are not in the same network anyways.

This all requires of course that both partners are online at the same time.
Later versions might make it possible to have a third party instance that acts
as intermediate cache. This would then resemble something like ``ownCloud`` does.

.. _about_names:

About names
~~~~~~~~~~~

You might already have wondered what those names that you pass on ``init`` are
and what they are for. ``brig`` does not impose any strict format on the
username. So any of these are valid usernames:

- ``alice``
- ``alice@wonderland.lit``
- ``alice@wonderland.lit/rabbithole``
- ``alice/rabbithole``

It's however recomended to choose a name that is formatted like
a `XMPP/Jabber-ID`_. Those IDs can look like plain emails, but can
optionally have a »resource« part as suffix (separated by a »/« like
``ovaloffice``). The advantage of having a username in this form is
locabillity: ``brig`` can find users with the same domain - which is useful for
e.g. companies with many users.

.. _`XMPP/Jabber-ID`: https://de.wikipedia.org/wiki/Jabber_Identifier

.. note::

    The domain part of the email does not need to be a valid domain,
    but later releases might add email based authentication schemes
    which will require a valid domain in the username.

Having a resource part is optional, but can help if you have several instances
of ``brig`` on your machines. i.e. one username could be
``alice@wonderland.org/desktop`` and the other ``alice@wonderland.org/laptop``.

.. note::

    The same name can be taken by more than one node. That's a result of the
    distributed nature of ``brig`` since there is no central part that can
    register all usernames persistently. This of course opens space for
    attackers: A malicious person can take the same username as your friend
    - but luckily he can't take over his fingerprint.

    ``brig`` does therefore not use the name to authenticate a user. This is done
    by the *fingerprint*, which is explained later. Think of the name
    as a human readable »DNS«-name for fingerprints for now.

Syncing
-------

Finally there. Let's recap what we've done so far:

- Create a repository (``brig init <name>``) - needs to be done only once.
- Find & add remotes (``brig remote add``) - needs to be done once for each peer.
- Add some files (``brig stage <path>``) - needs to be done as much as you like to.

Only thing left to do now is using ``brig diff`` and ``brig sync``.
First, let's check what changes ``bob`` has and how it will change our files:

.. code-block:: bash

    $ brig diff bob
    •
    ├── _ hello.world
    ├── + election
    └── README.md ⇄ README.md

``brig`` does not support showing what changed *in* a file, but it supports
how the file itself changed. For this we record the following type of changes:

====== ====================================================================
Symbol Description
====== ====================================================================
``+``  The file was added on the remote side.
``-``  The file was removed on the remote side.
``_``  The file is missing on the remote side (e.g. we added it)
``→``  The file was moved to a new location.
``*``  This file was ignored because we chose to due to our settings.
``⇄``  Both sides have changes, but they can be merged.
``⚡``  Both sides have changes, but they conflict.
====== ====================================================================

So in the above output we can tell that *Bob* added the directory
``/election``, but does not posess the ``/hello.world`` file. He also
apparently modified ``README.md``, but since we did not, it's safe for us to
take his changes. If we sync now we will get this directory from him:

.. code-block:: bash

    $ brig sync bob
    $ brig ls
    SIZE   MODTIME          PATH          PIN
    443 B  Dec 27 14:44:44  /README.md     🖈
    12 B   Dec 27 15:14:16  /hello.world   🖈
    32 GB  Dec 27 15:14:16  /election      🖈

You might notice that the ``sync`` step was kind of fast for 32 GB. This is
because ``sync`` *does not transfer actual data*. It only transferred the
metadata, while the actual ``data`` will only be loaded when required. This
also means that your data does not need to reside on the same device on which
you are using ``brig``. You could have one instance on your always online
server, while you use only tiny parts of it on your small netbook.

Where is the data then? Thanks to ``ipfs`` it can be transferred from anywhere,
but usually nodes that already downloaded the file from the origin. This is
another advantage of a distributed approach: The original node does not need to
be online as long as some other node also has your file stored. Note that your
node will not pro-actively gather data you won't use. It simply might cache
data longer than necessary.

How are the files secure then if they essentially could be everywhere?
Every file is encrypted by ``brig`` before giving it to ``ipfs``. The key is part
of the metadata and will be used to decrypt the file again on the receiver's end.

Pinning
-------

How do we control then what files are stored locally and what not? By *pinning*
each file or directory you want to keep always. Files you add explicitly are
pinned by default and also files that were synced to you. Only old versions of
a file are by default unpinned.

``brig`` knows of two types of pins: **Explicit** and **implicit**. When a file
or directory is being pinned by ``brig pin``, we call this an explicit pin,
since the user decided he wants to keep that file. When you update a file
locally, ``brig`` will unpin the old version and pin the new content
*implicitly*. In the command line output, explicit pins are always shows
magenta, while implicit pins are shown as implicit.

.. todo::

    Explain the implications of pinning when syncing files and other operations like reset.

If you never pin something explicitly, only the newest version of all files
will be stored locally. If you decide that you need older versions, you can pin
them explictly, so brig cannot unpin them implicily. For this you should also
look into the ``brig pin set`` and ``brig pin clear`` commands, which are
similar to ``brig pin add`` and ``brig pin rm`` but can operate on whole commit
ranges.

Once ``brig gc`` is being run, all files that are not pinned (explicit or
implcit) are being deleted from local storage. However, those files can be
still retrieved by other nodes that store the respective content.

Version control
---------------

One key feature of ``brig`` over other synchronisation tools is the handy
version control you can have. It will feel familiar to ``git`` users, but a few
concepts are different.

Key concepts
~~~~~~~~~~~~

This is written from the perspective of a ``git`` user:

* You can »snapshot« your current repository by creating a commit (``brig commit``)
* There are no detailed »diffs« between two files. Only a mix of the following state changes:

   - *added:* The file was added in this commit.
   - *moved:* The file was moved in this commit.
   - *removed:* The file was removed in this commit.
   - *modified:* The file's content was changed in this commit.

* A change is only recorded between individual commits. Changes in-between are
  not recorded.
* There are no branches. Every user has a linear list of commits. The choice
  not to have branches is on purpose, since they tend to bring greate
  complexity to both implementation and user-interface.
* Since there are no branches, there is no way to go back into history. You can
  however checkout previous files.
* You can tag individual commits. There are three pre-defined tags:

    - *STAGE*: The current, not yet finalized commit. Constantly changing.
    - *HEAD*: The last finished commit.
    - *INIT*: The first commit that was made.

* When synchronizing files with somebody, a merge commit is automatically created.
  It contains a special marker to indicate with whom, at what time and what state we merged with.
  On the next sync, commits before this merge will automatically be ignored.

Individual commands
~~~~~~~~~~~~~~~~~~~

* ``brig commit``: Create a new commit, possibly with a message that describes what happened in the commit.

* ``brig log``: Show a list of all commits, starting from the newest one.

  .. code-block:: bash

      $ brig log
      SEfXUBDu4J Dec 20 00:06:43 • (curr)
      SEfXUEVczh Dec 20 00:06:43 Added initial README.md (head)
      SEfXUEru1p Dec 20 00:06:43 initial commit (init)

* ``brig tag``: Tag a commit with a user defined name. This is helpful for
  remembering special commits like »homework-finale«.
* ``brig history``: Show the list of changes made to this file between commits.
* ``brig reset``: Checkout a whole commit or bring a single file or directory
  to the state of an old commit. In contrast to ``git``, checking out an old
  state works not by »jumpinp back«, but by setting the current commit
  (``STAGE``) to the contents of the old commit. It's a rather cheap operation
  therefore.
* ``brig diff / status``: Show the difference (i.e. what files were added/removed/moved/clashed)
* ``brig become``: View the files of a person we synced with.
