brig - decentralized & secure synchronization
=============================================

.. image:: _static/logo.png
   :width: 50%
   :align: center

What is ``brig``?
-----------------

``brig`` is a distributed & secure file synchronization tool with version
control. It is based on ``ipfs``, written in Go and will feel familiar to
``git`` users. Think of it as a swiss army knife for file synchronization or as
a distributed alternative to Dropbox.

**Key feature highlights:**

* Encryption of data in rest and transport, plus optional compression on the fly.
* Simplified ``git`` version control only limited by your storage space.
* Synchronization algorithm that can handle moved files and empty directories and files.
* Your data does not need to be stored on the device you are currently using.
* FUSE filesystem that feels like a normal sync folder.
* No central server at all. Still, central architectures can be build with ``brig``.
* Gateway to share normal HTTP/S links with other users.
* Auto-updating facility that will sync on any change.
* Completely free software under the terms of the ``AGPL``.
* Simple user identification and discovery.

Please refer to the :ref:`features-page` for more details. If you want a visual
hint how ``brig`` looks on the commandline, refer to the :ref:`quickstart`.

What is ``brig`` not?
---------------------

``brig`` tries to focus on being up conceptually simple, by hiding a lot of
complicated details regarding storage and security. Therefore the end result is
hopefully easy and pleasant to use, while being secure by default. Since
``brig`` is a »general purpose« tool for file synchronization it of course
cannot excel in all areas. It won't replace high performance network file
systems and should not be used when you are in need of high throughput - at
least not at the moment.

Current Status
--------------

**This software is in active development and probably not suited for production
use yet!** But to get it in a stable state, it is **essential** that people
play around with it. Consider this is as an open beta phase. Also don't take
anything granted for now, everything might change wildly before version ``1.0.0``.

With that being said, ``brig`` is near a somewhat usable state where you can play around
with it quite well. All aforementioned features do work, besides possibly being
a little harder to use than ideally possible. A lot of work is currently going into
stabilizing the current feature set.

At this moment ``brig`` is **only tested on Linux**. Porting and testing
efforts are welcome. Other platforms should be able to compile, but there are
currently not guarantees that it will work.

.. toctree::
   :maxdepth: 2
   :caption: Installation:

   installation.rst

.. toctree::
   :maxdepth: 2
   :caption: User manual

   tutorial/intro.rst
   tutorial/init.rst
   tutorial/coreutils.rst
   tutorial/mounts.rst
   tutorial/remotes.rst
   tutorial/pinning.rst
   tutorial/vcs.rst
   tutorial/gateway.rst
   tutorial/config.rst

.. toctree::
   :maxdepth: 2
   :caption: Additional resources

   quickstart.rst
   faq.rst
   features.rst
   comparison.rst

.. toctree::
   :maxdepth: 2
   :caption: Development

   roadmap.rst
   contributing.rst
