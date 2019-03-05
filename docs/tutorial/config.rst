.. _configurations:

Configuration
-------------

As mentioned earlier, we can use the built-in configuration system to configure many aspects
of ``brig`` functionality to our liking. Every config entry of ``brig`` consists of 4 values:

* Key - always a dotted, hierarchical path like ``fs.sync.ignore_moved``.
* Value - some value that is validated depending on the key.
* Default - The default value.
* Documentation - A short description of what this entry can do for you.
* Needs restart - A boolean indicating whether you have to restart the service to take effect.

When you type ``brig cfg`` you will see all keys with the aforementioned entries:

.. code-block:: bash

    $ brig config ls
    [...]
    fs.sync.ignore_moved: false (default)
      Default:       false
      Documentation: Do not move what the remote moved
      Needs restart: no
    [...]

Additionally, we support of course the usual operations:

.. code-block:: bash

    $ brig config get repo.password_command
    pass brig/repo/password
    $ brig config set repo.password_command "pass brig/repo/my-password"

Profiles
~~~~~~~~

.. todo:: Implement configuration profiles.
