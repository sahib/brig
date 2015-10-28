=====
USAGE
=====

Usage:

    brig SUBCOMMAND [OPTIONS …]

brig is a distributed file synchronization tool based on IFPS and XMPP. 
Every file repository (called a "port") is assigned to a JabberID like this:

    sahib@jabber.nullcat.de/laptop


REPOSITORIY COMMANDS:

    brig init  <PATH> [<JID>]      Initialize an empty port with no files at <PATH>
    brig clone <JID>               Clone an existing port fully or shallow to <PATH>
    brig open  <PATH>              Open an encrypted port. Asks for passphrase.
    brig close <PATH>              Closes an encrypted port.

DAEMON COMMANDS:

    brig watch <PATH> [--pause]    Watch a port for changes and add them automatically.
    brig daemon                    Start a communication daemon manually.
    brig sync [--peek] [-p <JID>]  Start a synchronization manually or look what would happen.
    brig push [--peek] [-p <JID>]  Push last added files to network (do not pull)
    brig pull [--peek] [-p <JID>]  Pull changes from peers (no push)

XMPP HELPER COMMANDS:

    brig discover                  Search network for potential peers (via Zeroconf locally).
    brig friends                   List all reachable and offline peers ("Buddy list")
    brig auth <JID> [--qa]         Send auth request to (potential) peer at <JID>
    brig ban <JID>                 Discontinue friendship with <JID>
    brig prio <JID> <N>            Set priority of peer to <N>

WORKING DIR COMMANDS:

    brig status                    Give a overview of brig's current state.
    brig add <FILE> [-p <JID>]     Make <FILE> managed by brig.
    brig copies <FILE> <N>         Keep at least <N> copies of <FILE>
    brig find <PATTERN>            Find filenames locally and in the net.
    brig rm <FILE>                 Puts copy of <FILE> in the trash bin or removes it directly.

DATA INTEGRITY COMMANDS:

    brig lock|unlock               Disallow or allow local or remote modifications of the port.
    brig verify [--fix]            Verify, and possibly fix, broken files.

REVISION COMMANDS:

    brig log <FILE>                Show all known versions of this file.
    brig checkout <FILE> <HASH>    Checkout old version of this file, if available.

SECURITY COMMANDS:

    brig yubi                      Manage yubikeys.
    brig key                       Manage your PGP key.

MISC COMMANDS:

    brig config <KEY> <VAL>        Set a config key. 
    brig config <KEY>              Get a config key.
    brig config -l                 List all available keys.
    brig update                    Try to securely update brig.
    brig help                      Show detailed help.

Files that match the patterns in .brignore files are not watched.

Config Values
=============

Similar to ``git`` there is a global configuration, where also all
repositories on the device is stored.

- Node Type: [archive, backup, desktop, checkout, hold]

    - archive  => revision control to certain depth, autosync to archive.
    - backup   => revision control with depth of 1.
    - client   => No revision control [default]
    - checkout => No autosync, only checkout certain files
    - hold     => Hold repository that removes file after certain time.

- Merge Priorities

    - Give certain peers a priority. In case of merge conflicts highest ranking
      will be kept.
    - Repositories with same bare jabber id get higher priority by default.
    - Maybe introduce a merge bin with files that need manual review?

- ...
