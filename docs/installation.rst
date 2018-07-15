Installation
------------

At the time of writing, there are no pre-compiled binaries. So you gonna have
to compile ``brig`` yourself - but don't worry that is quite easy:


Step 0: Installing Go
~~~~~~~~~~~~~~~~~~~~~

This is only required if you don't have ``Go`` installed.

.. warning::

    ``brig`` only works with a newer version of Go (>= 1.9).
    The version in your package manager might be too outdated,
    if you're on e.g. Debian. Make sure it's up to date!


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

You also need to have ``git`` installed for the next step.

.. todo:: Describe *how* for different distributions.

Step 1: Compile & Install ``brig``
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

This step requires that you have set ``GOPATH``, as discussed
in the previous section.

.. code:: bash

    $ go get -d -v -u github.com/sahib/brig  # Download the sources.
    $ cd $GOPATH/src/github.com/sahib/brig   # Go to the source directory.
    $ make                                   # Build the software.
    $ sudo make install                      # Install it system-wide (optional)

All dependencies of brig are downloaded for you during the first step.
Execution might take a few minutes though.

.. note::

    For the curious: Why the Makefile?

    In theory it's also possible to install ``brig`` via ``go get`` only, but
    there is to straight-forward way to set the git revision in the binary.
    Thus the Makefile.

Step 2: Test if the installation is working
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code:: bash

    $ brig help

If above command prints out documentation on how to use the program's
commandline switches then the installation worked. Congratulations!

-----

Continue with :ref:`getting_started` or directly go to :ref:`quickstart` if you
do not want to hear the details.
