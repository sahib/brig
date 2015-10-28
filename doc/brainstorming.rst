=============
BRAINSTORMING
=============

Technology
==========

IPFS
----

IPFS is short for the InterPlanetary FileSystem. It is a fast and efficient p2p
network. We use IPFS to form little sub-nets that can share files each other.
With IPFS we do not need to re-invent a lot of basic boilerplate features.
Also, by using IPFS we're not restricted to local network, but can also use the
whole IPFS network if necessary.

XMPP
----

XMPP (also called Jabber) is used as existing infrastructure to easily pair
devices with IDs. In the case of XMPP the ID is a JabberID like this:

    sahib@nullcat.de/laptop
    sahib@nullcat.de/desktop

The part before the slash is called a bare id, which is meant to be unique 
for your devices. Together with the resource behind the slash, the bare id forms
the full id which refers to a single device.

Zeroconf
--------

Easy decovery of other peers in the local network. The software needs to act as
client and server. This needs an Zeroconf server (e.g. Avahi) to run on all
sides to work.

inotify
-------

Watches a directory for modifications. Enables brig to
automatically commit 

This is not portable to windows? (or some other unices)


btrfs, zfs, rsync algorithm?
----------------------------

For archive nodes, old snapshots up to a certain depth would be valuable.
There are several ways to achieve this, for example using btrfs snapshots.
A more portable and self-contained alternative would be using the rsync
using the rsync algorithm to create incremental layers.

Terms
=====

Repository (Port)
-----------------

A repository is a folder with files and some special semantic in it. It can be
shared over several peers, either with full or partial content. A repository
does not necessarily sync all of it's contents. Like with ``git`` the
synchronization might be triggered automatically, or the directory might be
watched with ``inotify`` for autosync.

Open questions:

- Repo structure? Hidden directory? .brignore files?
- More than one repository possible with ipfs? Probably, how?

A file index is associated to each repository, which is shared fully with each
peer.

Peers (ships)
-------------

Other peers in the network you are authenticated too. These are either close
peers (same bare id as you or explicitly trusted) or friend peers.
Every peer will be added as ipfs bootstrap peer.
To be added as peers, other devices need to be authenticated.

Open questions:

- How does the auth work? OTR? Using the ipfs PGP key?

Security
========

Authentication
--------------

- Question/Answer?
- Verify public key?

File transfer
-------------

- Share one-time-keys over xmpp and encrypt files before sending with ipfs?

Libraries
=========

XMPP
----

go-xmpp2
