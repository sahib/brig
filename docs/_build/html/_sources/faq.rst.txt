Frequently Asked Questions
==========================

General questions
-----------------

1. Why is the software named ``brig``?
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

It is named after the ship with the same name.
When we named it, we thought it's a good name for the following reason:

- A ``brig`` is a very lightweight and fast ship.
- It was commonly used to transport small amount of goods.
- A ship operates on streams (sorry 😛)
- The name is short and somewhat similar to ``git``.
- It gives you a few nautical metaphors and a logo for free.
- Words like »bright«, »brigade« and many others start with it

Truth be told, only half of the two name givers thought it's a good name, but
I still kinda like it.

2. Who develops it?
~~~~~~~~~~~~~~~~~~~

Although this documentation sometimes speaks of »we«, the only developer is
currently `Chris Pahl <https://github.com/sahib>`_. He writes it entirely in
his free time, mostly during commuting with the train.

Technical questions
-------------------

1. How is the encryption working?
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

A stream is chunked into equal sized blocks that are encrypted in GCM mode
using AES-256. Additionally ChaCha20 (with Poly1305) is currently supported but
it might be removed soon. The overall file format is somewhat similar to NaCL
secretboxes, but it is more tailored to supporting efficient seeking.

The current default is ``ChaCha20``, although machines with the ``aes-ni``
instruction set might yield significant higher throughput. The source of the
`encryption layer can be found here <https://github.com/sahib/brig/tree/master/catfs/mio/encrypt>`_.
Here's a basic overview over the format:

.. image:: _static/format-encryption.svg
    :width: 66%
    :align: center

The key of each file is currently being derived from the content hash of the
file (See also `Convergent Encryption
<https://en.wikipedia.org/wiki/Convergent_encryption>`_). If the content
changes later, the key does not change since the key is only generated once
during the first staging of the file.

Please refer to the implementation for all implementation details for now. No
security audits of the implementation have been done yet, therefore I'd
appreciate every pair of eyes. Especially while everything is still in flux and
won't harm any users.

2. Is there compression implemented?
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

Yes. The compression is being done before encryption and is only enabled if the
file looks compression-worthy. The »worthiness« is determined by looking at its
header to guess a mime-type. Depending on the mime-type either ``snappy`` or
``lz4`` is selected or no compression is added at all.

The source of the `compression layer can be found here <https://github.com/sahib/brig/tree/master/catfs/mio/compress>`_. Here's
a basic overview over the format:

.. image:: _static/format-compression.svg
    :width: 66%
    :align: center

3. What hash algorithms are used?
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

Two algorithms are used:

* ``SHA256`` is used by ``IPFS`` for every backend hash.
* ``SHA3-256`` is used as general purpose hash for everything ``brig`` internal
  (Content and Tree hash).

Each hash is encoded as `multihash
<https://github.com/multiformats/multihash>`_. For output purposes this
representation is encoded additionally in ``base58``. Therefore, all hashes
that start with ``W1`` are ``sha3-256`` hashes while the ones starting with
``Qm`` are ``sha256`` hashes. Keep in mind that ``base58`` is case-sensitive.

4. What kind of deduplication is currently used?
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

It is currently only possible to deduplicate between individual versions of a file.
And there also only the portion before the modification.

``IPFS`` implements deduplication, but it is circumvented by encrypting blocks
before giving them over to the backend. Implementing a more proper and informed
deduplication is one of the long term goals, which require more thorough
interaction with ``IPFS``. It is also possible to do some basic deduplication
purely on ``brig`` side since we have more info on the file than ``IPFS`` has.

5. How fast is the I/O when using ``brig``?
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

Here are some rather outdated graphs where you can get a rough feeling how fast
it can be. There are a few rules of thumb with mostly obvious content:

* It it goes over the network, it's the network speed plus a smaller constant overhead.
* If it comes over FUSE, it is quite a bit slower than over ``brig cat``.
* If you do not use compression, writing and reading will be faster.

The graphs below only measure in-memory performance compared to a ``dd`` like
speed (see the »baseline« line).

.. image:: _static/movie_read.svg
    :width: 66%

.. image:: _static/movie_write.svg
    :width: 66%

Your mileage may vary and you better do your own benchmarks for now.

.. todo::

    Explain/Update those graphs.
