.. BADGES:

.. API documentation:
.. image:: https://godoc.org/github.com/sahib/config?status.svg
   :target: https://godoc.org/github.com/sahib/config

.. Test status via Travis:
.. image:: https://img.shields.io/travis/sahib/config/master.svg?style=flat
   :target: https://travis-ci.org/sahib/config

.. Issue tracker:
.. image:: https://img.shields.io/github/issues/sahib/config.svg?style=flat
   :target: https://github.com/sahib/config/issues

.. Release overview:
.. image:: https://img.shields.io/github/release/sahib/config.svg?style=flat
   :target: https://github.com/sahib/config/releases

.. Download count:
.. image:: https://img.shields.io/github/downloads/sahib/config/latest/total.svg
   :target: https://github.com/sahib/config/releases/latest

.. GPL tag:
.. image:: http://img.shields.io/badge/license-GPLv3-4AC51C.svg?style=flat
   :target: https://www.gnu.org/licenses/quick-guide-gplv3.html.en

``config``
==========

Go package for loading typesafe program configuration with validation and migration.

Motivation
----------

There are already multiple packages available for loading application
configuration from a file. Viper_ is just one of the more popular ones. They
mostly focus on being convinient to use and to support lots of different
configuration sources (environment variables, network sources, files...). I
found none though that focuses on loading configuration and validating the
incoming values. Some libraries supported default values, but all of them
allowed »free form« configs -- i.e. defining keys that are not available in the
program without any warning. This usually leads to a lot of user errors, wrong
configuration and/or misbehaving programs. Config libraries should be very
strict in what they accept and allow for validation and migration of the config
when the layout changes -- this is why I wrote this package. If you didn't
notice yet: The design is somewhat opinionated and not tailored to cover all
possible use cases or to be overly convinient in exchange for power.

*Note:* The API is influenced by Python's ConfigObj_, but ``specs.cfg`` is part of the program.

.. _Viper: https://github.com/spf13/viper
.. _ConfigObj: http://configobj.readthedocs.io/en/latest/configobj.html

Features
--------

**Validated:** All configuration keys that are possible to set are defined with
a normal type in a ``.go`` file as part of your application. Those values are
also taken over as defaults, if they are not explicitly overwritten. This
ensures that the program always have sane values to work with. Every key can be
associated with a validation func, which can be used to implement further
validation (i.e. allow a string key to only have certain enumeration values).

**Versioned:** Every config starts with a version of zero. If the application
owning the config needs to change the layout, it can register a migration
function to do this once an old configuration is loaded. This frees you from worrying
about breaking changes in the config.

**Typesafety**: There is no stringification of values or other surprises (like
in *ConfigObj*). Every configuration key has exactly one value, directly
defined in Go's type system.

**Change Notification and instant reloading:** The application can reload the
configuration anytime and also register a func that will be called when a
certain key changes. This allows longer running daemon processes to react
instantly on config changes, if possible.

**Built-in Documentation:** You can write down documentation for your configuration
as part of the defaults definition, including a hint if this key needs a restart of
the application to take effect.

**Support for multiple formats:** Only YAML is supported by default, since it
suffices in the vast majority of use cases. If you need to, you can define your
own ``Encoder`` and ``Decoder`` to fit your specific usecase.

**Support for sub-sections:** Sub-sections of a config can be used like a
regular config object. Tip: Define your configuration hierarchy like the
package structure of your program. That way you can pass sub-section config to
those packages and you can be sure that they can only change the keys they are
responsible for.

**Support for placeholder sections:** By using the special section name ``__many__``
you can have several sections that all follow the same layout, but are allowed to be
named differently.

**Native support for slices:** All native types (``string``, ``int``, ``float`` and ``bool``)
can be packed into lists and easily set and accessed with special API for them.

**Merging:** Several configs from several sources can be merged. This might be
useful e.g. when there are certain global defaults, that are overwritten with local
defaults which are again merged with user defined settings.

**Reset to defaults:** Any part of the config can be reset to defaults at any time.

Examples
--------

- `Basic example.`_

.. _`Basic example.`: https://github.com/sahib/config/blob/master/example_test.go#L51

- `Migration example.`_

.. _`Migration example.`: https://github.com/sahib/config/blob/master/example_test.go#L127

If the validation was not succesful, you can either error out directly or continue with defaults as fallback:

.. code-block:: go

    // Somewhere in your init code:
    cfg, err := config.Open(config.NewYamlDecoder(fd), defaults)
    if err != nil {
        log.Errorf("Failed to user config. Continuing with defaults.")
        cfg, err = config.Open(nil, defaults)
        if err != nil {
            // Something is really wrong. The defaults are probably wrong.
            // This is a programmer error and should be catched early though.
            log.Fatalf("Failed to load default config: %v", err)
        }
    }


LICENSE
-------

`config` is licensed under the conditions of the `GPLv3
<https://www.gnu.org/licenses/quick-guide-gplv3.html.en>`_. See the
``COPYING``. distributed along the source for details.

Author
------

Christopher <sahib_> Pahl 2018

.. _sahib: https://www.github.com/sahib

----

Originally developed as part of »brig« (https://github.com/sahib/brig)
