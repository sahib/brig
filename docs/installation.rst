Installation
------------

At the time of writing, there are no pre-compiled binaries. So you gonna have
to compile ``brig`` yourself - but don't worry that is quite easy. We do not
have many dependencies, you only need two things: The programming language *Go*
and the version control system ``git``.

Step 0: Installing Go
~~~~~~~~~~~~~~~~~~~~~

This is only required if you don't already have ``Go`` installed.
Please consult your package manager for that.

.. warning::

    ``brig`` only works with a newer version of Go (>= 1.9).
    The version in your package manager might be too outdated,
    if you're on e.g. Debian. Make sure it's rather up to date!


If you did not do that, you gonna need to install ``Go``. `Refere here
<https://golang.org/doc/install>`_ for possible ways of doing so. Remember to
set the ``GOPATH`` environment variable to a place where you'd like to have
your ``.go`` sources being placed. For example you can put this in your
``.bashrc``:

.. code:: bash

    # Place the go sources in a "go" directory inside your home directory:
    export GOPATH=~/go
    # This is needed for the go toolchain:
    export GOBIN="$GOPATH/bin"
    # Make sure that our shell finds the go binaries:
    export PATH="$GOPATH/bin:$PATH"

By choosing to have the ``GOPATH`` in your home directory you're not required
to have ``sudo`` permissions later on. You also need to have ``git``
`installed <https://git-scm.com/download/linux>`_ for the next step.

Step 1: Compile & Install ``brig``
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

This step requires setting ``GOPATH``, as discussed in the previous section.

.. code:: bash

    $ go get -d -v -u github.com/sahib/brig  # Download the sources.
    $ cd $GOPATH/src/github.com/sahib/brig   # Go to the source directory.
    $ make                                   # Build the software.

All dependencies of brig are downloaded for you during the first step.
Execution might take a few minutes though because all of ``brig`` is being
compiled during the ``make`` step.

If you cannot or want to install ``git`` for some reason, you can `manually
download a zip <https://github.com/sahib/brig/archive/master.zip>`_ from GitHub
and place its contents into ``$GOPATH/src/github.com/sahib/brig``. In this
case, you can skip the ``go get`` step.

Step 2: Test if the installation is working
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

If everything worked, there will be a ``brig`` binary in ``$GOBIN``.

.. code:: bash

    $ brig help

If above command prints out documentation on how to use the program's
commandline switches then the installation worked. Happy shipping!

-----

Continue with :ref:`getting_started` or directly go to :ref:`quickstart` if you
just need a refresh on the details.
