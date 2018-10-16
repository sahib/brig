brig - decentralized & secure file synchronization
==================================================

.. image:: _static/logo.png
   :width: 50%
   :align: center

``brig`` is a distributed & secure file synchronization tool with version control.
It is based on ``ipfs``, written in Go and will feel familiar to ``git`` users.

**Key feature highlights:**

* Encryption of data in rest and transport + compression on the fly.
* Simplified ``git`` version control.
* Sync algorithm that can handle moved files and empty directories and files.
* Your data does not need to be stored on the device you are currently using.
* FUSE filesystem that feels like a normal (sync) folder.
* No central server at all. Still, central architectures can be build with ``brig``.
* Simple user identification and discovery with users that look like email addresses.

``brig`` tries to focus on being up conceptually simple, by hiding a lot of
complicated details regarding storage and security. Therefore I hope the end
result is easy and pleasant to use, while being secure by default.
Since ``brig`` is a "general purpose" tool for file synchronization it of course
cannot excel in all areas. This is especially true for efficiency, which is
sometimes sacrificed to get the balance of usability and security right.

Current Status
--------------

**This software is in active development and not suited for production use yet!**
But to get there I need people that try it and verify their workflows with ``brig``.

Apart from that, ``brig`` is near a somewhat usable state where you can play around
with it quite well. All aforementioned features do work, besides possibly being
a little harder to use than ideally possible. A lot of work is currently going into
stabilizing the current feature set.

At this moment ``brig`` is **only available for Linux**. Porting efforts are
welcome though. The only blocker is *FUSE* which would be needed to either
replaced on other platforms or disabled.

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
