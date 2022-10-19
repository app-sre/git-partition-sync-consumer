package pkg

import (
	"fmt"
	"net/url"
	"os/exec"
)

// Push local repos to remotes
func PushLatest(gitlabUsername, gitlabToken string, archives []ArchiveInfo) error {
	for _, archive := range archives {
		authURL, err := formatAuthURL(archive.RemoteURL, gitlabUsername, gitlabToken)
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
func formatAuthURL(gitURL, gitlabUsername, gitlabToken string) (string, error) {
	u, err := url.Parse(gitURL)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s://%s:%s@%s%s",
		u.Scheme,
		gitlabUsername,
		gitlabToken,
		u.Host,
		u.Path,
	), nil
}
