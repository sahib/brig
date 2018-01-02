Introduction
============

Installation
------------

At the time of writing, there are no pre-compiled binaries.
So you gonna have to opt-out and compile ``brig`` yourself,
but don't worry that is quite easy:

.. note::

    ``brig`` currently was only tested on Linux with the most
    Go version (1.9.2 currently)


Step 0: Installing Go
~~~~~~~~~~~~~~~~~~~~~

This is only required if you don't have ``Go`` installed.

If you did not do that, you gonna need to install ``Go``. `Refere here
<https://golang.org/doc/install>`_ for possible ways of doing so. Remember to
set the ``GOPATH`` to a place where you'd like to have your ``.go`` files being
placed. For example you can put this in your ``.bashrc``:

.. code:: bash

    # Place the go sources in a "go" directory inside your home directory:
    export GOPATH=~/go
    export PATH="$GOPATH/bin:$PATH"

By choosing to have the ``GOPATH`` in your home directory you're not required
to have ``sudo`` permissions later on.

You also need to have ``git`` and ``hg`` installed for the next step.

.. todo:: describe how.

Step 1: Compile & Install ``brig``
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code:: bash

    $ go get -u github.com/sahib/brig


This will download the complete source code of ``brig`` (and all of it's
dependencies) and compile them to a binary right on. Execution might take a few
minutes.

Step 2: Test if the installation is working
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code:: bash

    $ brig help

If above command prints out documentation on how to use the program's
commandline switches then the installation worked. Congratulations!

-----

Continue with :ref:`getting_started` or directly go to :ref:`quickstart` if you
do not want to hear the details.
