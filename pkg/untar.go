package pkg

import (
	"archive/tar"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	pushPkg "github.com/app-sre/git-partition-sync-producer/pkg"
)

type UntarInfo struct {
	DirPath      string
	LocalBranch  string
	RemoteGroup  string
	RemoteName   string
	RemoteBranch string
	ShortSHA     string
}

const UNTAR_DIRECTORY = "untarred-repos"

// "untar" the content of decrypted s3 objects
// each directory is created at current working dir with name of object key
// adaption of: https://medium.com/@skdomino/taring-untaring-files-in-go-6b07cf56bc07
func (d *Downloader) extract(decrypted []*DecryptedObject) ([]*UntarInfo, error) {

	err := d.clean(UNTAR_DIRECTORY)
	if err != nil {
		return nil, err
	}

	archives := []*UntarInfo{}

	// "untar" each s3 object's body and output to directory
	// each dir is name of the s3 object's key (this is base64 encoded still)
	for _, dec := range decrypted {
		b64GitInfo := strings.SplitN(dec.Key, ".", 2)[0]
		path := filepath.Join(d.workdir, UNTAR_DIRECTORY, b64GitInfo)

		if err := Untar(dec.DecryptedTar, path); err != nil {
			return nil, err
		}

		// track newly untarred repo for future git operations
		a := &UntarInfo{DirPath: path}
		err = d.extractGitRemote(a, dec.Key)
		if err != nil {
			return nil, err
		}
		archives = append(archives, a)
	}
	return archives, nil
}

func Untar(tarred io.Reader, path string) error {
	gzr, err := gzip.NewReader(tarred)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	err = os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return err
	}

untar:
	for {
		header, err := tr.Next()
		switch {
		case err == io.EOF:
			break untar
		case err != nil:
			return err
		case header == nil:
			continue
		}

		target := filepath.Join(path, header.Name)
		switch header.Typeflag {
		case tar.TypeDir:
			if _, err := os.Stat(target); err != nil {
				if err := os.MkdirAll(target, 0755); err != nil {
					return err
				}
			}
		case tar.TypeReg:
			// if file is encountered before parent dir has been created
			dir := filepath.Dir(target)
			if _, err := os.Stat(dir); err != nil {
				if err := os.MkdirAll(dir, 0755); err != nil {
					return err
				}
			}
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			// copy over contents
			if _, err := io.Copy(f, tr); err != nil {
				return err
			}
			// manually close here after each file operation; defering would cause each file close
			// to wait until all operations have completed.
			f.Close()
		default:
			fmt.Println(header.Typeflag)
			return errors.New(
				fmt.Sprintf("Unable to untar `%s` object. Encountered unsupported type", target),
			)
		}
	}
	return nil
}

// decodes an s3 object key and extracts the gitlab remote target information
func (d *Downloader) extractGitRemote(a *UntarInfo, encodedKey string) error {
	// remove file extension before attempting decode
	// extension is .tar.age, split at first occurrence of .
	encodedGitInfo := strings.SplitN(encodedKey, ".", 2)[0]
	decodedBytes, err := base64.StdEncoding.DecodeString(encodedGitInfo)
	if err != nil {
		return err
	}

	// unmarshal decoded key (json) into struct defined by gitlab-sync-s3-push
	var jsonKey pushPkg.DecodedKey
	err = json.Unmarshal(decodedBytes, &jsonKey)
	if err != nil {
		return err
	}

	a.LocalBranch = jsonKey.LocalBranch
	a.RemoteGroup = jsonKey.Group
	a.RemoteName = jsonKey.ProjectName
	a.RemoteBranch = jsonKey.RemoteBranch
	a.ShortSHA = jsonKey.CommitSHA[:7] // only take 7 characters of sha
	return nil
}
