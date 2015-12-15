package im

import (
	"bytes"
	"encoding/xml"

	"github.com/tsuibin/goxmpp2/xmpp"
)

func truncate(a string, l int) string {
	if len(a) > l {
		return a[:l] + "..." + a[len(a)-l:]
	}

	return a
}

func createMessage(from, to xmpp.JID, text string) *xmpp.Message {
	xmsg := &xmpp.Message{}
	xmsg.From = from
	xmsg.To = to
	xmsg.Id = xmpp.NextId()

	xmsg.Type = "chat"
	xmsg.Lang = "en"
	xmsg.Body = []xmpp.Text{
		{
			XMLName:  xml.Name{Local: "body"},
			Chardata: text,
		},
	}

	return xmsg
}

func joinBodies(msg *xmpp.Message) []byte {
	if msg == nil {
		return nil
	}

	buf := &bytes.Buffer{}
	for _, field := range msg.Body {
		buf.Write([]byte(field.Chardata))
	}

	return buf.Bytes()
}
