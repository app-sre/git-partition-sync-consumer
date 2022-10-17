package pkg

import (
	"fmt"
	"net/url"
	"os/exec"
	"time"
)

// shield your eyes
// TODO: replace this evil with https://github.com/go-git/go-git
// Push repos to new branches on target projects
func PushLatest(gitlabUsername, gitlabToken string, archives []ArchiveInfo) error {
	for _, archive := range archives {
		authURL, err := formatAuthURL(archive.RemoteURL, gitlabUsername, gitlabToken)
		if err != nil {
			return err
		}
		branchName := fmt.Sprintf("commercial_%s_%d", archive.ShortSHA, time.Now().UnixMilli())

		args := []string{
			"-c",
			fmt.Sprintf("%s && %s && %s && %s && %s && %s",
				fmt.Sprint("git init --initial-branch=main"),
				fmt.Sprintf("git checkout -b %s", branchName),
				fmt.Sprintf("git remote add fedramp %s", authURL),
				fmt.Sprint("git add ."),
				fmt.Sprintf("git commit -m 'Sync to version %s'", archive.ShortSHA),
				fmt.Sprintf("git push -u fedramp %s", branchName),
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
