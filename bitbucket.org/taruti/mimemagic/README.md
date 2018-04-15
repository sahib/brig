# Mimemagic - Detect mime-types in Go

    import "bitbucket.org/taruti/mimemagic"

Import the library

    mimemagic.Match(myguess string, startoffile []byte) string

The API:
* myguess is a guess of the mimetype that is checked first before other types are checked. It may also be "".
* startoffile contains the beginning of the file - e.g. giving it 1024 bytes from the start of the file works fine.
* the result is the guessed mimetype or "" if not known.

