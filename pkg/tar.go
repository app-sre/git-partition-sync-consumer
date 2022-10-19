package pkg

import (
	"archive/tar"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type ArchiveInfo struct {
	DirPath      string // absolute // START HERE: need to account for repos being nested under /clone_workdir // how do we set specific repo names?
	RemoteURL    string
	RemoteBranch string
	ShortSHA     string
}

// unzip the content of decrypted s3 objects
// each directory is created at current working dir with name of object key
// adaption of: https://medium.com/@skdomino/taring-untaring-files-in-go-6b07cf56bc07
func Extract(decrypted []DecryptedObject) ([]ArchiveInfo, error) {
	working, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	err = clear(working)
	if err != nil {
		return nil, err
	}

	archives := []ArchiveInfo{}

	// "untar" each s3 object's body and output to directory
	// each dir is name of the s3 object's key (this is base64 encoded still)
	for _, dec := range decrypted {
		tr := tar.NewReader(dec.Archive)
		path := filepath.Join(working, "repos", dec.Key)
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
		err = extractGitRemote(&a, dec.Key)
		if err != nil {
			return nil, err
		}
		archives = append(archives, a)
	}
	return archives, nil
}

// decodes an s3 object key and extracts the gitlab remote url
// expected format of decoded s3 key is destination_url/commit_sha/branch
func extractGitRemote(a *ArchiveInfo, encodedKey string) error {
	// original object keys are base64 encoded and appended with file extension
	encodedKeySegments := strings.SplitN(encodedKey, ".", 2)
	decodedKeyBytes, err := base64.StdEncoding.DecodeString(encodedKeySegments[0])
	if err != nil {
		return err
	}
	decodedKey := string(decodedKeyBytes)
	decodedKeySegments := strings.Split(decodedKey, "/")
	url := fmt.Sprintf("%s.git",
		// remove the trailing commit sha and branch
		strings.Join(decodedKeySegments[:len(decodedKeySegments)-2], "/"),
	)
	if err != nil {
		return err
	}
	a.RemoteURL = url
	a.RemoteBranch = decodedKeySegments[len(decodedKeySegments)-1]
	a.ShortSHA = decodedKeySegments[len(decodedKeySegments)-2][:7] // only take 7 characters of sha
	return nil
}

// ensure clean directory directory before unzipping
func clear(working string) error {
	cmd := exec.Command("rm", "-rf", "repos/")
	cmd.Dir = working
	err := cmd.Run()
	if err != nil {
		return err
	}
	return err
}
