1. Init
=======

Usage is very close to ``git``.

.. code-block:: bash

    $ mkdir repo && cd repo
    # Create a new repository in here:
    $ brig init sahib@wald.de/laptop
    $ ls
    # Test if it works:
    $ brig cat README.md

Explain:

- Files created in init.

2. Adding files
===============

Explain why it's "stage" not "add"

.. code-block:: bash

    $ brig stage ~/music.mp3
    $ brig ls
    $ brig tree
    $ brig cat music.mp3 | mpv -

Explain:

- Path names.

3. Coreutils
============

Explain reflinks.

.. code-block:: bash

    $ brig mkdir sub
    $ brig cp music.mp3 sub
    $ brig tree

    $ brig info README.md
    $ brig edti README.md
    $ brig edit README.md
    $ brig info README.md

    $ ipfs cat <hash>
    -> garbled bullshit.

Explain:

- Most coreutils are available.
- Hash changed after editing.

4. Mounting
===========

.. code-block:: bash

    $ brig ls
    $ mkdir /tmp/mount
    $ ls /tmp/mount  # Empty.
    $ brig mount /tmp/mount
    $ nautilus /tmp/mount
    $ vi /tmp/mount/new-file
    $ brig ls
    $ cp ~/bbb.mkv /tmp/mount
    $ mpv /tmp/mount/bbb.mkv

Problem: Performance? Pinning again?

5. Commits
==========

.. code-block:: bash

    $ brig log
    $ brig commit -m 'Added darth vader'
    $ brig log
    $ brig edit README.md
    $ brig mv sub/music.mp3 sub/else.mp3
    $ brig diff   # Should print mergeable and moved file.

Problem: Diff shows mv order wrong way?

6. History
==========

.. code-block:: bash

    # Little different than git.
    $ brig history new-file
    $ brig edit new-file
    $ brig commit -m 'edited new-file'
    $ brig reset HEAD^ new-file
    $ brig cat new-file

BUG: brig reset is doing bullshit.

7. Discovery & Remotes
======================

.. code-block:: bash

    $ bob-brig ls
    $ brig whoami
    $ brig net locate alice
    $ brig remote add <name> <hash>
    $ brig remote ls
    $ brig remote edit

Docker bereits laufen lassen?

8 Sync & Diff
=============

.. code-block:: bash

    $ brig remote ls
    $ brig diff alice
    $ brig sync alice
    $ brig log
    $ brig ls

9 Pinning
=========

.. code-block:: bash

    $ brig pin rm <path-of-bob>
    $ brig gc
    $ <close bob docker>
    $ brig cat <path>
    ...blocks...

10 Misc
=======

.. code-block:: bash

    $ brig bug
    $ brig docs
