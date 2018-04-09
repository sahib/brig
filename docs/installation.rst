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

.. todo:: describe how for different distributions.

Step 1: Compile & Install ``brig``
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code:: bash

    $ git clone --recursive https://github.com/sahib/brig
    $ cd brig && make
    $ sudo make install

This will download the complete source code of ``brig`` (and all of it's
dependencies). The second step compiles the binary. Execution might take a few
minutes though. The third step will install the binary to ``/usr/local/bin/brig`` -
you can of course copy it to another path.


.. note::

    In theory it's also possible to install ``brig`` via ``go get``, but there
    is currently a misbehaviour in ``go get`` that prevents proper downloading
    of submodules. Also, the resulting binary lacks the git revision in the
    help output, which is why we discourage this method.

Step 2: Test if the installation is working
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code:: bash

    $ brig help

If above command prints out documentation on how to use the program's
commandline switches then the installation worked. Congratulations!

-----

Continue with :ref:`getting_started` or directly go to :ref:`quickstart` if you
do not want to hear the details.
