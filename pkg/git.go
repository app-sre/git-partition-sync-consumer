package pkg

import (
	"fmt"
	"os/exec"
)

// shield your eyes
// TODO: replace this evil with https://github.com/go-git/go-git
// Push repos to new branches on target projects
func PushLatest(gitlabUsername, gitlabToken string, archives []ArchiveInfo) error {
	for _, archive := range archives {
		args := []string{
			"-c",
			fmt.Sprintf("%s && %s && %s && %s && %s && %s",
				fmt.Sprint("git init --initial-branch=main"),
				fmt.Sprintf("git checkout -b commercial_%s", archive.ShortSHA),
				fmt.Sprintf("git remote add fedramp %s", archive.RemoteSshUri),
				fmt.Sprint("git add ."),
				fmt.Sprintf("git commit -m 'Sync to version %s'", archive.ShortSHA),
				fmt.Sprintf("git push -u fedramp commercial_%s", archive.ShortSHA),
			),
		}
		cmd := exec.Command("/bin/sh", args...)
		cmd.Dir = archive.DirPath
		err := cmd.Run()
		if err != nil {
			return err
		}
	}
	return nil
}
