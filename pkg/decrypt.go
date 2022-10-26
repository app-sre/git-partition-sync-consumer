package pkg

import (
	"io"

	"filippo.io/age"
)

type EncryptedObject interface {
	Key() string
	Reader() io.ReadCloser
}

type DecryptedObject struct {
	Key          string
	DecryptedTar io.Reader
	err          error
}

func (d *Downloader) decryptBundles(objects []EncryptedObject) ([]*DecryptedObject, error) {
	decrypted := []*DecryptedObject{}

	identity, err := age.ParseX25519Identity(d.privateKey)
	if err != nil {
		return nil, err
	}

	for _, encObj := range objects {
		dec, err := age.Decrypt(encObj.Reader(), identity)
		if err != nil {
			return nil, err
		}

		decrypted = append(decrypted, &DecryptedObject{
			Key:          encObj.Key(),
			DecryptedTar: dec,
		})
	}

	return decrypted, nil
}
