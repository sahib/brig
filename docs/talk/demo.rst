0. Preparation
==============

- Windows: Chrome Incognito (slides, presenter console), Monitor Settings, Terminal (docker, ipfs, hovercraft), Terminal (empty)

-----

- setxkbmap us && xmodmap ~/.xmodmaprc
- check sound.
- Check that docker is running.
- Check that no other brig instance is up.
- Check: /tmp/{repo,mount} is empty.
- Do a "bob-brig ls and bob-brig rmt ls" to do some pre-caching.
- Source autocompletion.

1. Init
=======

Usage is very close to ``git``.

.. code-block:: bash

    $ mkdir repo && cd repo
    # Create a new repository in here:
    # Command started einen daemon im Hintergrund!
    $ brig init alice@wonderland.de/laptop
    # Anschaut was brig so angestellt hat:
    $ ls
    # Dann schauen wir mal ob man die Datei ausgeben kann:
    $ brig cat README.md

2. Adding files
===============

.. code-block:: bash

    $ brig stage ~/music.mp3
    $ brig ls
    # Pfadnamen, virtueller root.
    $ brig tree
    $ brig cat music.mp3 | mpv -


3. Coreutils
============

.. code-block:: bash

    $ brig mkdir sub
    $ brig cp music.mp3 sub
    $ brig tree

    # ähnlich zu `stat` unter linux:
    $ brig info README.md
    $ brig edti README.md
    $ brig edit README.md
    # Hash hat sich nach Edit-Vorgang geändert:
    $ brig info README.md

    # Man kann sich ansehen was für daten ipfs dann speichert:
    $ ipfs cat <hash>
    -> garbled bullshit.

4. Mounting
===========

.. code-block:: bash

    $ mkdir /tmp/mount
    $ ls /tmp/mount  # Empty.
    $ brig mount /tmp/mount
    # Ta-da, alle dateien die man sonst so hat sind auch hier vorhanden:
    $ nautilus /tmp/mount
    # Man kann ganz normal dateien editieren:
    $ vi /tmp/mount/new-file
    $ brig ls
    # Noch nicht sehr performant, aber sowas geht schon:
    $ cp ~/rrd.mkv /tmp/mount
    $ mpv /tmp/mount/rrd.mkv

5. Commits
==========

.. code-block:: bash

    $ brig log
    $ brig diff
    $ brig commit -m 'Added darth vader'
    $ brig log
    $ brig edit README.md
    $ brig mv sub/music.mp3 sub/else.mp3
    $ brig diff   # Should print mergeable and moved file.

6. History
==========

(optional)

.. code-block:: bash

    # Etwas anders als git: kein diff an sich:
    $ brig history new-file
    $ brig edit new-file
    $ brig commit -m 'edited new-file'
    $ brig reset HEAD^ new-file
    $ brig cat new-file

7. Discovery & Remotes
======================

.. code-block:: bash

    # bob läuft in einem container auf dem gleichen computer:
    $ bob-brig ls
    $ brig whoami
    # Erst ausführen, dauert etwas:
    $ brig net locate bob
    $ brig remote add $(bob-brig whoami -f)
    $ bob-brig remote add $(brig whoami -f)
    $ brig remote ls
    $ brig remote edit

8 Sync & Diff
=============

.. code-block:: bash

    $ brig remote ls
    $ brig diff bob
    $ brig sync bob
    $ brig log
    $ brig ls

9 Pinning
=========

.. code-block:: bash

    $ brig pin rm <path-of-bob> # geht.
    $ brig gc
    $ brig cat <path>           # geht.
    $ <close bob docker>
    $ brig gc
    $ brig cat <path>
    ...blocks...

10 Misc
=======

.. code-block:: bash

    $ brig <tab>
    $ brig help stage
    $ brig docs
    $ brig bug
