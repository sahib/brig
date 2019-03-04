.. _configurations:

Configuration
-------------

Quite a few details can be configured in a different way to your liking. ``brig
config`` is the command that allows you to list, get and set individual
configuration values. Each config entry already brings some documentation that
tells you about its purpose:

.. code-block:: bash

    $ brig config ls
    [... output truncated ...]
    fs.sync.ignore_moved: false (default)
    Default:       false
    Documentation: Do not move what the remote moved
    Needs restart: no
    [... output truncated ...]
    $ brig config get repo.password_command
    pass brig/repo/password
    $ brig config set repo.password_command "pass brig/repo/my-password"
