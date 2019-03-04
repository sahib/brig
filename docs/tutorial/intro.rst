.. _getting_started:

Getting started
================

This guide will walk you through the steps of synchronizing your first files
over ``brig``. You will learn about the concepts behind it along the way.
Most of the steps here will include working in a terminal, since this is the primary
way to interact with ``brig``. Once setup you have to choice to use a browser application though.

Precursor: The help system
--------------------------

Before we dive in, we go over a few things that will make your life easier
along the way. ``brig`` has some built-in helpers to serve as support for your
memory.

Built-in reference documentation
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

Every command offers detailed built-in help, which you can view using the
``brig help`` command. This often usage examples too:

.. code-block:: bash

    $ brig help stage
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

       $ brig stage file.png                         # gets added as /file.png
       $ brig stage file.png /photos/me.png          # gets added as /photos/me.png
       $ cat file.png | brig stage --stdin /file.png # gets added as /file.png

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

After starting a new shell you should be able to autocomplete most commands.
Try this for example by typing ``brig remote <tab>``. Other shells are not
supported right now sadly.

Open the online documentation
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

By typing ``brig docs`` you'll get a tab opened in your default browser with this
domain loaded. Please stop typing ``brig documentation`` into Google.

Reporting bugs
~~~~~~~~~~~~~~~

If you need to report a bug you can use a built-in utility to do that. It will
gather all relevant information, create a report and open a tab with the
*GitHub* issue tracker in a browser for you. Only thing left for you is to fill
out some questions in the report and include anything you think is relevant.

.. code-block:: bash

    $ brig bug

To actually create the issue you sadly need an *GitHub* `account
<https://github.com/join>`_. If  you don't have internet or do not want to sign
up, you can still generate a bug report template via ``brig bug -s``.
