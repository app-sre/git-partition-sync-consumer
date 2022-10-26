package pkg

import (
	"fmt"
	"net/url"
	"os/exec"
)

// Push local repos to remotes
func (d *Downloader) pushLatest(archives []*UntarInfo) error {
	for _, archive := range archives {

		authURL, err := d.formatAuthURL(fmt.Sprintf("%s/%s", archive.RemoteGroup, archive.RemoteName))
		if err != nil {
			return err
		}

		args := []string{
			"-c",
			fmt.Sprintf("%s && %s",
				fmt.Sprintf("git remote add fedramp %s", authURL),
				fmt.Sprintf("git push -u fedramp %s", archive.RemoteBranch),
			),
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
