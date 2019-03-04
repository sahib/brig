Version control
---------------

One key feature of ``brig`` over other synchronisation tools is the built-in
and quite capable version control. If you already know ``git`` that's a plus
for this chapter since a lot of stuff will feel similar. This isn't a big
surprise, since ``brig`` implements something like ``git`` internally. Don't
worry, knowing ``git`` is however not needed at all for this chapter.

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
are used to. Together with pinning this offers a high degree of freedom on how
you can decide what repositories store what data. The price is that this
fine-tuned control can get a little annoying. Future versions of ``brig`` will
try to solve that.

For some more background, you can invoke ``brig info`` to see what metadata is
being saved per file version:

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
    Tree Hash     W1gX8NMQ9m8SBnjHRGtamRAjJewbnSgi6C1P7YEunfgTA3
    Content Hash  W1pzHcGbVpXaePa1XpehW4HGPatDUJs8zZzSRbpNCGbN2u
    Backend Hash  QmPvNjR1h56EFK1Sfb7vr7tFJ57A4JDJS9zwn7PeNbHCsK


Most of it should be no big surprise. It might be a small surprise that three
hashes are stored per file. The ``Backend Hash`` is really the link to the
actual data. If you'd type ``ipfs cat
QmPvNjR1h56EFK1Sfb7vr7tFJ57A4JDJS9zwn7PeNbHCsK`` you will get the encrypted
version of your file dumped to your terminal. The ``Content Hash`` is being
calculated before the encryption and is the same for two files with the same
content. The ``Tree Hash`` is a hash that uniquely identifies this specific
node for internal purposes. The ``Inode`` is a number that stays unique over
the lifetime of a file (including moves and removes). It is used mostly in the
FUSE filesystem.

Commits
~~~~~~~

Now that we know that only metadata is versioned, we have to ask »what is the
smallest unit of modification that can be saved?«. This smallest unit is
a commit. A commit can be seen as a snapshot of the whole repository.

The command ``brig log`` shows you a list of commits that were made already:

.. code-block:: bash

          -      Sun Oct 14 22:46:00 CEST 2018 • (curr)
    W1kAySD3aKLt Sun Oct 14 22:46:00 CEST 2018 user: Added ali-file (head)
    W1ocyBsS28SD Sun Oct 14 22:46:00 CEST 2018 user: Added initial README.md
    W1D9KsLNnAv4 Sun Oct 14 22:46:00 CEST 2018 initial commit (init)


Each commit is identified by a hash (e.g. ``W1kAySD3aKLt``) and records the
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
    W1hZoY7TrxyK Sun Oct 14 22:46:00 CEST 2018 user: better leave some bread crumbs (head)

This snapshot can be useful later if you decide to revert to a certain version.
The hash of the commit is of course hard to remember, so if you need it very often, you can
give it a tag yourself. Tags are similar to the names, ``curr``, ``head`` and ``init`` but
won't be changed by ``brig`` and won't move therefore:

.. code-block:: bash

    # instead of "W1hZoY7TrxyK" you also could use "head" here:
    $ brig tag W1hZoY7TrxyK breadcrumbs
    $ brig log | grep breadcrumbs
    $ W1hZoY7TrxyK Sun Oct 14 22:46:00 CEST 2018 user: better leave some bread crumbs (breadcrumbs, head)


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
    added   INIT  W1ocyBsS28SD    Oct 14 22:46:00
    $ brig mv README.md README_LATER.md
    $ brig hst README_LATER.md
    CHANGE  FROM  TO            HOW                           WHEN
    moved   HEAD  CURR          /README.md → /README_LATER.md Oct 15 00:27:37
    added   INIT  W1ocyBsS28SD                                Oct 14 22:46:0

As you can see, you will be shown one line per history entry. Each entry
denotes which commit the change was in. Some commits were nothing was changed
will be jumped over except if you pass ``--empty``.

Viewing differences
~~~~~~~~~~~~~~~~~~~

If you're interested what changed in a range of your own commits, you can use
the ``brig diff`` command as shown previously. The ``-s`` (``--self``) switch
says that we want to compare only two of our own commits (as opposed to
comparing with the commits of a remote).

.. code-block:: bash

    # Let's compare the commit hashes from above:
    $ brig diff -s W1hZoY7TrxyK W1kAySD3aKLt
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
    W1hZoY7TrxyK Sun Oct 14 22:46:00 CEST 2018 user: better leave some bread crumbs (breadcrumbs)


As you can see, we still have the previous commits. ``brig revert`` did one
thing more than restoring the state of ``init`` and put that result in
``curr``. This also means that you can't really *modify* history. But you can
revert it. Let's revert your complete wipe-out:

.. code-block:: bash

    # Reset to the state we had in »breadcrumbs«
    $ brig reset breadcrumbs


``brig reset`` cannot only restore old commits, but individual files and
directories:

.. code-block:: bash

    $ brig reset head^^ README.md

.. note::

    It is a good idea to do a ``brig commit`` before a ``brig reset``. Since it
    modifies ``curr`` you might loose uncommitted changes. It will warn you
    about that, but you can overwrite that warning with ``--force``. If you did
    a ``brig commit`` you can simply use ``brig reset head`` to go back to the
    last good state.


Nodes that were overwritten with ``brig reset`` will be unpinned (unless pinned
explicitly). Those nodes and their content will be garbage collected after some
time. The content may still be accessed through the use of other remotes
though.
