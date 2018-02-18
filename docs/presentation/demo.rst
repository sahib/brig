0. Install
===========

.. code-block:: bash

    $ go get github.com/sahib/brig

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

2. Adding files
===============

Explain why it's "stage" not "add"

.. code-block:: bash

    $ brig stage music.mp3
    $ brig ls
    $ brig tree
    $ brig cat music.mp3 | mpv -

Explain: Path names.

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

4. Commits
==========

.. code-block:: bash

    $ brig log
    $ brig commit -m 'Added darth vader'
    $ brig log
    $ brig edit README.md
    $ brig mv sub/music.mp3 sub/else.mp3
    $ brig diff   # Should print mergeable and moved file.

5. History
==========

.. code-block:: bash

    $ brig history README.md
