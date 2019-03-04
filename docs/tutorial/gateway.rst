Screenshots
-----------

Here are some screenshots of the gateway

.. image:: ../_static/gateway-login.png
    :alt: Gateway login screen
    :width: 66%

.. image:: ../_static/gateway-files.png
    :alt: Gateway files view
    :width: 66%

.. image:: ../_static/gateway-changelog.png
    :alt: Gateway changelog view
    :width: 66%

.. image:: ../_static/gateway-trashbin.png
    :alt: Gateway trashbin view
    :width: 66%

.. image:: ../_static/gateway-remotes.png
    :alt: Gateway remotes view
    :width: 66%

.. image:: ../_static/gateway-add-remote.png
    :alt: Gateway add remote view
    :width: 66%


Using the gateway
-----------------

Many users will not run ``brig``. Chances are, that you still want to send or
present them your files without too much hassle. ``brig`` features a *Gateway*
to HTTP(S), which comes particularly handy if you happen to run a public
server and/or want to provide a GUI to your users.

Before you do anything, you need to a »user« to your gateway. This user is different
than remotes and describes what credentials can be used to access the gateway.
You can add add a new user like this:

.. code-block:: bash

    $ brig gateway user add admin my-password
    # or shorter:
    # brig gw u a admin my-password
    $ brig gateway user list
    NAME  FOLDERS
    admin /

The gateway is disabled by default. If you want to start it, use this command:

.. code-block:: bash

    $ brig gateway start

Without further configuration, this will create a HTTP (**not** HTTPS!) server
on port ``5000``, which can be used already. If you access it under ``http://localhost:5000``
you will see a login mask where you can log yourself in with the credentials you used earlier.

If you'd like to use another port than ``5000``, you can do so by setting the
respective config key:

.. code-block:: bash

    $ brig cfg set gateway.port 7777

.. note::

    You can always check the status of the gateway:

    .. code-block:: bash

        $ brig gateway status

    This will also print helpful diagnostics if something might be wrong.

The gateway can be stopped anytime with the following command. It tries to still
serve all open requests, so that no connections are dropped:

.. code-block:: bash

    $ brig gateway stop

.. note::

    If you want to forward the gateway to the outside, but do not own
    a dedicated server, you can forward port 5000 to your computer. With this
    setup you should also get a certficate which in turn requires a DNS name.
    An easy way to get one is to use dynamic DNS.

There is also a small helper that will print you a nice hyperlink to a certain
file called ``brig gateway url``:

.. code-block:: bash

    $ brig gateway url README.md
    http://localhost:5000/get/README.md


Securing access
~~~~~~~~~~~~~~~

You probably do not want to offer your files to everyone that have a link.
Therefore you can restrict access to a few folders (``/public`` for example)
and require a user to authenticate himself with a user and password upon access.

By default all files are accessible. You can change this by changing the config:

.. code-block:: bash

    $ brig cfg set gateway.folders /public

Now only the files in ``/public`` (and including ``/public`` itself) are
accessible from the gateway. If you want to add basic HTTP authentication:


.. code-block:: bash

    $ brig cfg set gateway.auth.enabled true
    $ brig cfg set gateway.auth.user <user>
    $ brig cfg set gateway.auth.pass <pass>

If you use authentication, it is strongly recommended to enable HTTPS.
Otherwise the password will be transmitted in clear text.


Running the gateway with HTTPS
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

The gateway has built-in support for `LetsEncrypt <https://letsencrypt.org/>`_.
If the gateway is reachable under a DNS name, it is straightforward to get
a TLS certificate for it. In total there are three methods:

**Method one: Automatic:** This works by telling the gateway the domain name.
Since the retrieval process for getting a certificate involves binding on port 80,
you need to prepare the brig binary to allow that without running as root:

.. code-block:: bash

    # You need to restart the brig daemon for that.
    # Every next brig command will restart it.
    $ brig daemon quit
    $ sudo setcap CAP_NET_BIND_SERVICE=+ep $(which brig)

Afterwards you can set the domain in the config. If the gateway is already running,
it will restart immediately.

.. code-block:: bash

    $ brig cfg set gateway.cert.domain your.domain.org

You can check after a few seconds if it worked by checking if the ``certfile`` and ``keyfile``
was set:

.. code-block:: bash

    $ brig cfg get gateway.cert.certfile
    /home/user/.cache/brig/your.domain.org_cert.pem
    $ brig cfg get gateway.cert.keyfile
    /home/user/.cache/brig/your.domain.org_key.pem
    $ curl -i https://your.domain.org:5000
    HTTP/2 200
    vary: Accept-Encoding
    content-type: text/plain; charset=utf-8
    content-length: 38
    date: Wed, 05 Dec 2018 11:53:57 GMT

    This brig gateway seems to be working.

This method has the advantage that the certificate can be updated automatically
before it expires.

**Method two: Half-Automated:**

If the above did not work for whatever reasons, you can try to get a certificate manually.
There is a built-in helper called ``brig gateway cert`` that can help you doing that:

.. code-block:: bash

    $ brig gateway cert your.domain.org
    You are not root. We need root rights to bind to port 80.
    I will re-execute this command for you as:
    $ sudo brig gateway cert nwzmlh4iouqikobq.myfritz.net --cache-dir /home/sahib/.cache/brig

    A certificate was downloaded successfully.
    Successfully set the gateway config to use the certificate.
    Note that you have to re-run this command every 90 days currently.

If successful, this command will set the ``certfile`` and ``keyfile`` config
values for you. You can test if the change worked by doing the same procedure
as in *method one*. Sadly, you have to re-execute once the certificate expires.

**Method three: Manual:**

If you already own a certificate you can make the gateway use it by setting the path
to the public certificate and the private key file:

.. code-block:: bash

    $ brig cfg set gateway.cert.certfile /path/to/cert.pem
    $ brig cfg set gateway.cert.keyfile /path/to/key.pem

If you do not own a certificate yet, but want to setup an automated way to
download one for usages outside of brig, you should look into
`certbot <https://certbot.eff.org/docs/>`_.

Redirecting HTTP traffic
~~~~~~~~~~~~~~~~~~~~~~~~

This section only applies to you if you choose **method one** from above and
want to run the gateway on port 80 (http) and port 443 (https). This has the
advantage that a user does not need to specify the port in a gateway URL have
which looks a little bit less *»scary«*. With this setup all traffic on port 80
will be redirected directly to port 443.

.. code-block:: bash

    $ brig cfg set gateway.port 443
    $ brig cfg set gateway.cert.redirect.enabled true
    $ brig cfg set gateway.cert.redirect.http_port 80
