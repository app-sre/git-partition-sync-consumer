package pkg

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
)

// Push local repos to remotes
func (d *Downloader) pushLatest(archives []*UntarInfo) error {
	// include command to trust internal git server certificate if set
	caPath := os.Getenv("INTERNAL_GIT_CA_PATH")

	for _, archive := range archives {
		authURL, err := d.formatAuthURL(fmt.Sprintf("%s/%s", archive.RemoteGroup, archive.RemoteName))
		if err != nil {
			return err
		}

		var args []string
		if len(caPath) > 0 {
			args = []string{
				"-c",
				fmt.Sprintf("%s && %s && %s",
					// this config must be set per repo (cannot be done once with --global flag)
					// due to permission constraint when attempting to edit root gitconfig
					fmt.Sprintf("git config http.sslCAInfo %s", caPath),
					fmt.Sprintf("git remote add fedramp %s", authURL),
					fmt.Sprintf("git push -u fedramp %s --force", archive.RemoteBranch),
				),
			}
		} else {
			args = []string{
				"-c",
				fmt.Sprintf("%s && %s",
					fmt.Sprintf("git remote add fedramp %s", authURL),
					fmt.Sprintf("git push -u fedramp %s --force", archive.RemoteBranch),
				),
			}
		}

		cmd := exec.Command("/bin/sh", args...)
		cmd.Dir = archive.DirPath
		err = cmd.Run()
		if err != nil {
			return err
		}
	}
	return nil
}

// returns git user-auth format of remote url
func (d *Downloader) formatAuthURL(pid string) (string, error) {
	projectURL := fmt.Sprintf("%s/%s", d.glBaseURL, pid)
	parsedURL, err := url.Parse(projectURL)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s://%s:%s@%s%s.git",
		parsedURL.Scheme,
		d.glUsername,
		d.glToken,
		parsedURL.Host,
		parsedURL.Path,
	), nil
}
