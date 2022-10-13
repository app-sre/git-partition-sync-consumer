package pkg

import (
	"bytes"
	"io"
	"os"
	"sync"

	"golang.org/x/crypto/openpgp"
)

type GpgHelper struct {
	Entities openpgp.EntityList
}

// NewGpgHelper initializes a GpgHelper object and configures the private key
func NewGpgHelper(path, passphrase string) (GpgHelper, error) {
	helper := GpgHelper{}

	// credit: https://gist.github.com/stuart-warren/93750a142d3de4e8fdd2
	// open private key file
	buffer, err := os.Open(path)
	if err != nil {
		return helper, err
	}
	defer buffer.Close()

	// retrieve entities from private key
	// private key should be armored
	entityList, err := openpgp.ReadArmoredKeyRing(buffer)
	if err != nil {
		return helper, err
	}
	entity := entityList[0]

	// read private key using passphrase
	passphraseBytes := []byte(passphrase)
	entity.PrivateKey.Decrypt(passphraseBytes)
	for _, subkey := range entity.Subkeys {
		subkey.PrivateKey.Decrypt(passphraseBytes)
	}

	helper.Entities = entityList
	return helper, nil
}

type EncryptedObject interface {
	Key() string
	Reader() io.ReadCloser
}

type DecryptedObject struct {
	Key     string
	Archive io.Reader
	err     error
}

// accepts list of encrypted interface objects and concurrently decrypts the objects
func (g *GpgHelper) DecryptBundles(objects []EncryptedObject) ([]DecryptedObject, error) {
	result := []DecryptedObject{}

	var wg sync.WaitGroup
	ch := make(chan DecryptedObject)

	for _, obj := range objects {
		wg.Add(1)
		go g.decrypt(&wg, ch, obj)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	for dec := range ch {
		if dec.err != nil {
			return nil, dec.err
		}
		result = append(result, dec)
	}

	return result, nil
}

// goroutine func that sends a DecryptedObject on the channel
// accepts an s3object and decrypts the content using gpg private key
// expects encrypted content to not be armored
func (g *GpgHelper) decrypt(wg *sync.WaitGroup, ch chan<- DecryptedObject, object EncryptedObject) {
	defer wg.Done()
	dec := DecryptedObject{Key: object.Key()}

	// read s3 object contents
	objBytes, err := io.ReadAll(object.Reader())
	if err != nil {
		dec.err = err
		ch <- dec
		return
	}

	// decrypt body(repo archive) using private key
	details, err := openpgp.ReadMessage(bytes.NewBuffer(objBytes), g.Entities, nil, nil)
	if err != nil {
		dec.err = err
		ch <- dec
		return
	}

	dec.Archive = details.UnverifiedBody
	ch <- dec
}
