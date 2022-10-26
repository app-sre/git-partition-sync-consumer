package pkg

import (
	"archive/tar"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	pushPkg "github.com/dwelch0/gitlab-sync-s3-push/pkg"
)

type ArchiveInfo struct {
	DirPath      string
	RemoteGroup  string
	RemoteName   string
	RemoteBranch string
	ShortSHA     string
}

// unzip the content of decrypted s3 objects
// each directory is created at current working dir with name of object key
// adaption of: https://medium.com/@skdomino/taring-untaring-files-in-go-6b07cf56bc07
func (d *Downloader) extract(decrypted []*DecryptedObject) ([]ArchiveInfo, error) {
	const UNTAR_DIRECTORY = "untarred-repos"

	err := d.clean(UNTAR_DIRECTORY)
	if err != nil {
		return nil, err
	}

	archives := []ArchiveInfo{}

	// "untar" each s3 object's body and output to directory
	// each dir is name of the s3 object's key (this is base64 encoded still)
	for _, dec := range decrypted {
		tr := tar.NewReader(dec.DecryptedTar)
		path := filepath.Join(d.workdir, UNTAR_DIRECTORY, dec.Key)
		err = os.MkdirAll(path, os.ModePerm)
		if err != nil {
			return nil, err
		}
		first := true
		var gitNestedPath string
	untar:
		for {
			header, err := tr.Next()
			switch {
			case err == io.EOF:
				break untar
			case err != nil:
				return nil, err
			case header == nil:
				continue
			}

			target := filepath.Join(path, header.Name)
			switch header.Typeflag {
			case tar.TypeDir:
				if _, err := os.Stat(target); err != nil {
					if err := os.MkdirAll(target, 0755); err != nil {
						return nil, err
					}
				}
				// the first dir encountered during unzip is the git repo's root
				// store this for later reference by git.go when attempting to cd to each git repo
				if first {
					gitNestedPath = header.Name
					first = false
				}
			case tar.TypeReg:
				f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
				if err != nil {
					return nil, err
				}
				// copy over contents
				if _, err := io.Copy(f, tr); err != nil {
					return nil, err
				}
				// manually close here after each file operation; defering would cause each file close
				// to wait until all operations have completed.
				f.Close()
			default:
				fmt.Println(header.Typeflag)
				return nil, errors.New(
					fmt.Sprintf("Unable to untar %s object. Encountered unsupported type", dec.Key),
				)
			}
		}
		// track newly unzipped repo for future git operations
		a := ArchiveInfo{DirPath: filepath.Join(path, gitNestedPath)}
		err = d.extractGitRemote(&a, dec.Key)
		if err != nil {
			return nil, err
		}
		archives = append(archives, a)
	}
	return archives, nil
}

// decodes an s3 object key and extracts the gitlab remote target information
func (d *Downloader) extractGitRemote(a *ArchiveInfo, encodedKey string) error {
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

	a.RemoteGroup = jsonKey.Group
	a.RemoteName = jsonKey.ProjectName
	a.RemoteBranch = jsonKey.Branch
	a.ShortSHA = jsonKey.CommitSHA[:7] // only take 7 characters of sha
	return nil
}
