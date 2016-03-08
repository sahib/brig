package im

import (
	"crypto/tls"
	"fmt"
	"os"
	"path/filepath"

	"github.com/disorganizer/brig/util"
	"github.com/tsuibin/goxmpp2/xmpp"
)

var (
	AliceJid = xmpp.JID("alice@jabber.nullcat.de/laptop")
	BobJid   = xmpp.JID("bob@jabber.nullcat.de/desktop")
	AlicePwd = "ThiuJ9wesh"
	BobPwd   = "eecot3oXan"
)

func NewDummyClient(user xmpp.JID, password string) (cl *Client, err error) {
	budPath := filepath.Join(os.TempDir(), "otr.bud."+user.Node())
	keyPath := filepath.Join(os.TempDir(), "otr.key."+user.Node())

	if err := util.Touch(budPath); err != nil {
		return nil, err
	}

	// Disable tls checks for the sake of getting travis to run:
	tlsConfig := tls.Config{InsecureSkipVerify: true}

	client, err := NewClient(&Config{
		Jid:             user,
		Password:        password,
		TLSConfig:       tlsConfig,
		KeyPath:         keyPath,
		FingerprintPath: budPath,
	})

	if err != nil {
		return nil, err
	}

	return client, nil
}

// MakeBuddies creates a fingerprint database for each client
// and fills it with the fingerprint of each other client.
//
// This is useful for tests, since a valid fingerprint exchange
// needs to have happened before a succesful connect.
func MakeBuddies(clients ...*Client) (paths []string, err error) {
	for userIdx, userClient := range clients {
		user := userClient.C.Jid
		path := filepath.Join(os.TempDir(), "otr.bud."+user.Node())
		fd, fdErr := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)

		if fdErr != nil {
			err = fdErr
			return
		}

		paths = append(paths, path)

		for buddyIdx, buddyClient := range clients {
			buddy := buddyClient.C.Jid
			if userIdx == buddyIdx {
				continue
			}

			_, fpErr := fmt.Fprintf(fd, "%s: %s\n", string(buddy), buddyClient.Fingerprint())
			if err != nil && fpErr != nil {
				err = fpErr
				break
			}
		}

		err = fd.Close()
		if err != nil {
			break
		}
	}

	return
}
