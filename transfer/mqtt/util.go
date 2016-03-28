package mqtt

import (
	log "github.com/Sirupsen/logrus"
	"github.com/bkaradzic/go-lz4"
	"github.com/disorganizer/brig/transfer"
	"github.com/gogo/protobuf/proto"
)

func payloadToProto(msg proto.Message, data []byte, authMgr transfer.AuthManager) error {
	decryptData, err := authMgr.Decrypt(data)
	if err != nil {
		return err
	}

	decompData, err := lz4.Decode(decryptData, decryptData)
	if err != nil {
		return err
	}

	if err := proto.Unmarshal(decompData, msg); err != nil {
		return err
	}

	return nil
}

func protoToPayload(msg proto.Message, authMgr transfer.AuthManager) ([]byte, error) {
	data, err := proto.Marshal(msg)
	if err != nil {
		return nil, err
	}

	compData, err := lz4.Encode(data, data)
	if err != nil {
		return nil, err
	}

	log.Debugf(
		"Compressed message from %.1fKB to %1.fKB (%.1f%%)",
		float64(len(data))/1024,
		float64(len(compData))/1024,
		float64(len(compData))/float64(len(data))*100,
	)

	return authMgr.Encrypt(compData)
}
