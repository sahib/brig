brig - decentralized & secure file synchronization
==================================================

.. image:: _static/logo.png
   :width: 50%
   :align: center

``brig`` is a distributed & secure file synchronization tool with version control.
It is based on ``ipfs``, written in Go and will feel familiar to ``git`` users.

**Key feature highlights:**

* Encryption of data in rest and transport, plus optional compression on the fly.
* Simplified ``git`` version control only limited by your storage space.
* Synchronization algorithm that can handle moved files and empty directories and files.
* Your data does not need to be stored on the device you are currently using.
* FUSE filesystem that feels like a normal sync folder.
* No central server at all. Still, central architectures can be build with ``brig``.
* Simple user identification and discovery.
* Gateway to share normal HTTP/S links with other users.
* Auto-updating facility that will sync on any change.
* Completely free software under the terms of the ``AGPL``.

``brig`` tries to focus on being up conceptually simple, by hiding a lot of
complicated details regarding storage and security. Therefore the end result is
hopefully easy and pleasant to use, while being secure by default. Since
``brig`` is a »general purpose« tool for file synchronization it of course
cannot excel in all areas. It won't replace high performance network file
systems.

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
efforts are welcome. The only incompatible feature so far is *FUSE* which would
be needed to either disabled or replaced on other platforms.

.. toctree::
   :maxdepth: 2
   :caption: User documentation:

   installation.rst
   introduction.rst
   quickstart.rst
   faq.rst

.. toctree::
   :maxdepth: 2
   :caption: Complementary

   comparison.rst
   roadmap.rst
   contributing.rst
