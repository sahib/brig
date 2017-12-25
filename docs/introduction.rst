.. _getting_started:

Getting started
================

This guide will walk you through the steps of synchronizing your first files
over ``brig``. It's hand's on, so make sure to open a terminal.
We'll explain all import concepts along the way.

Creating a repository
---------------------

You need a central place where your files are stored and ``brig`` calls this
place the *repository*. Note that this is not directly comparable to what
other tools calls the *Sync folder*. Rather think of it as the ``.git`` folder
of a ``git``-repository: A place where all internal state, data and metadata
of ``brig`` is stored.

By creating a new repository you also generate your identity, under which
your buddies can later find and authenticate you.

Enough of the grey theory, let's get started:

.. code-block:: bash

    $ brig init donald@whitehouse.gov/ovaloffice

The name you specified after the ``init`` is the name that will be shown
to other users and by which you are searchable in the network.
See :ref:`about_names` for more details on the subject.

You will be asked to enter a new password. *»Why«* you ask? This password is
used to store your data in an encrypted manner on your harddisk. This is
especially important if you think about creating the repository on a portable
media (e.g. usb sticks). If you still choose to disable this security feature
you're free to do so by passing ``-x`` directly before the ``init`` subcommand.

TODO: Write about:

- Sync folder in other tools.
- You can have more than one.

Adding files
------------

Phew, that was a lot of text, but there was not any action yet.
Let's change that by adding some files to ``brig``:

.. code-block:: bash

    $ brig stage 

TODO: Write about:

- Two path namespaces (external/internal)


Remotes
-------

About names
~~~~~~~~~~~

TODO: Move this section down a bit.

You surely have noticed that you had to specify a name during ``init``.
Since ``brig`` is built on-top of ``ipfs``, all users can find each other
and sync files among them. The name is used as a human readable token
to (hopefully) uniquely identify a single user.

.. note::

    ``brig`` does not use the name to authenticate a user. This is done
    by the *fingerprint*, which is explained later. Think of the name
    as a »DNS«-name for fingerprints.

Names can be used to locate other users:

.. code-block:: bash

    $ brig locate alice@wonderland.org
