# Change Log

All notable changes to this project will be documented in this file.

The format follows [keepachangelog.com]. Please stick to it.

## [0.4.1 Capricious Clownfish] -- unreleased

A smaller release with some bug fixes and a few new features. Also one bigger
stability and speed improvement. Thanks to everyone that gave feedback!

### Fixed

- Fix two badger db related crashes that lead to a crash in the daemon. One was
  related to having nested transactions, the one was related to having an open
  iterator while committing data to the database.
- Fix some dependencies that led to errors for some users (thanks @vasket)
- The gateway code now tries to reconnect the websocket whenever it was closed
  due to bad connectivity or similar issues. This led to a state where files
  were only updated after reloading the page.
- Several smaller fixes in the remotes view, i.e. the owner name was displayed
  wrong and most of the settings could not be set outside the test environment.
  Also the diff output was different in the UI and brig diff.
- We now error out early if e.g. »brig ls« was issued, but there is no repo.
  Before it tried to start a daemon and waited a long time before timing out.
- Made »brig mkdir« always prefix a »/« to a path which would lead to funny
  issues otherwise.

### Added

- Add a --offline flag to the following subcommands: ``cat``, ``tar``,
  ``mount`` and ``fstab add``. These flags will only output files that are
  locally cached and will not cause timeouts therefore. Trying other files will
  result in an error.
- »brig show« now outputs if a file/directory is locally cached. This is not
  the same as pinned, since you can pin a file but it might not be cached yet.
- Make the gateway host all of its JavaScript, fonts and CSS code itself by
  baking it into the binary. This will enable people running the gateway in
  environments where no internet connection is available to reach the CDN used
  before.
- Add the possibility to copy the fingerprint in the UI via a button click.
  Before the fingerprint was shown over two lines which made copying tricky.
- A PKGBUILD for ArchLinux was added, which builds ``brig`` from the
  ``develop`` branch. Thanks @vasket!

### Changed

- The ``brig remote ls`` command no longer does active I/O between nodes to check
  if a node is authenticated. Instead it relies on info from the peer server
  which can apply better caching. The peer server is also able to use information
  from dials and requests to/from other peers to update the ping information.
- Switch the internal checksum algorithm to ``blake2s-256`` from ``sha3-256``.
  This change was made for speed reasons and leads to a slightly different looking
  checksum format in the command line output. This change MIGHT lead to incompatibilities.
- Also swap ``scrypt`` with ``argon2`` for key derivation and lower the hashing settings
  until acceptable performance was achieved.
- Replace the Makefile with a magefile, i.e. a build script written in Go only which has
  no dependencies and can bootstrap itself.
- Include IPFS config output in »brig bug«.

### Removed

* The old Makefile was removed and replaced with a Go only solution.

## [0.4.0 Capricious Clownfish] -- 2019-03-19

It's only been a few months since the last release (December 2018), but there
are a ton of new features / general changes that total in about 15k added lines
of code. The biggest changes are definitely refactoring IPFS into its own
process and providing a nice UI written in Elm. But those are just two of the
biggest ones, see the full list below.

As always, ``brig`` is **always looking for contributors.** Anything from
feedback to pull requests is greatly appreciated.

### Fixed

- Many documentation fixes and updates.
- Gateway: Prefer server cipher suites over client's choice.
- Gateway: Make sure to enable timeouts.
- Bugfix in catfs that could lead to truncated file streams.
* Lower the memory hunger of BadgerDB.
* Fix a bug that stopped badger transactions when they got too big.

### Added

* The IPFS daemon does not live in the ``brig`` process itself anymore.
  It can now use any existing / running IPFS daemon. If ``ipfs`` is not installed,
  it will download a local copy and setup a repository in the default place.
  Notice that this is a completely backwards-incompatible change.

* New UI: The Gateway feature was greatly extended and an UI was developed that
  exposes many features in an easily usable way to people that are used to a
  Dropbox like interface. See
  [here](https://brig.readthedocs.io/en/develop/tutorial/gateway.html) for some
  screenshots of the UI and documentation on how to set it up. The gateway
  supports users with different roles (``admin``, ``editor``, ``collaborator``,
  ``viewer``, ``link-only``) and also supports logging as anonymous user (not by
  default!). You can also limit what users can see which folders.

* New event subsystems. This enables users to receive updates in "realtime"
  from other remotes. This is built on top of the experimental pubsub feature
  of IPFS and thus needs a daemon that was started with
  ``--enable-pubsub-experiment``. Users can decide to receive updates from
  a remote by issuing ``brig remote auto-update enable <remote name>``. [More
  details in the documentation](https://brig.readthedocs.io/en/develop/tutorial/remotes.html#automatic-updating).

* Change the way pinning works. ``brig`` will not unpin old versions anymore,
  but leave that to the [repinning settings](https://brig.readthedocs.io/en/develop/tutorial/pinning.html#repinning).
  This is an automatic process that will make sure to keep at least ``x``
  versions, unpin all versions greater than ``y`` and make sure that only a
  certain filesystem quota is used.

* New ``trash`` subcommand that makes it easy to show deleted files (``brig
  trash ls``) and undelete them again (``brig trash undelete <path>``).

* New ``brig push`` command to ask a remote to sync with us. For this to work
  the remote needs to allow this to us via ``brig remote auto-push enable <remote
  name>``. See also the
  [documentation](https://brig.readthedocs.io/en/develop/tutorial/remotes.html#pushing-changes).

* New way to handle conflicts: ``embrace`` will always pick the version of the remote you are syncing with.
  This is especially useful if you are building an archival node where you can push changes to.
  See also the [documentation](https://brig.readthedocs.io/en/develop/tutorial/remotes.html#conflicts).
  You can configure the conflict strategy now either globally, per remote or for a specific folder.

* Read only folders. Those are folders that can be shared with others, but when
  we synchronize with them, the folder is exempted from any modifications.

* Implement automated invocation of the garbage collector of IPFS. By default
  it is called once per hour and will clean up files that were unpinned. Note
  that this will also unpin files that are not owned by ``brig``! If you don't want this,
  you should use a separate IPFS instance for ``brig``.

* It's now possible to create ``.tar`` files that are filtered by certain patterns.
  This functionality is currently only exposed in the gateway, not in the command line.

* Easier debugging by having a ``pprof`` server open by default (until we
  consider the daemon to be stable enough to disable it by default). You can get
  a performance graph of the last 30s by issuing ``go tool pprof -web
  "http://localhost:$(brig d p)/debug/pprof/profile?seconds=30"``

* One way install script to easily get a ``brig`` binary in seconds on your computer:
  ``bash <(curl -s https://raw.githubusercontent.com/sahib/brig/master/scripts/install.sh)``

### Changed

* Starting with this release we will provide pre-compiled binaries for the most common platforms on the [release page](https://github.com/sahib/brig/releases).
* Introduce proper linting process (``make lint``)
* ``init`` will now set some IPFS config values that improve connectivity and performance
  of ``brig``. You can disable this via ``--no-ipfs-optimization``.
* Disable pre-caching by default due to extreme slow-ness.
* Migrate to ``go mod`` since we do not need to deal with ``gx`` packages anymore.
* There is no (broken) ``make install`` target anymore. Simply do ``make`` and
  ``sudo cp brig /usr/local/bin`` or wherever you want to put it.

### Removed

* A lot of old code that was there to support running IPFS inside the daemon process.
  As a side effect, ``brig`` is now much snappier.

## [0.3.0 Galloping Galapagos] -- 2018-12-07

### Fixed

- Compression guessing is now using Go's http.DetectContentType()

### Added

* New gateway subcommand and feature. Now files and directories can be easily
  shared to non-brig users via a normal webserver. Also includes easy https setup.

### Changed

### Removed

### Deprecated

## [0.2.0 Baffling Buck] -- 2018-11-21

### Fixed

All features mentioned in the documentation should work now.

### Added

Many new features, including password management, partial diffs and partial syncing.

### Changed

Many internal things. Too many to list in this early stage.

### Removed

Nothing substantial.

### Deprecated

Nothing.

## [0.1.0 Analphabetic Antelope] -- 2018-04-21

Initial release on the Linux Info Day 2018 in Augsburg.

[unreleased]: https://github.com/sahib/rmlint/compare/master...develop
[0.1.0]: https://github.com/sahib/brig/releases/tag/v0.1.0
[keepachangelog.com]: http://keepachangelog.com/
