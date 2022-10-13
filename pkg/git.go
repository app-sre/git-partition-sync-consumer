package pkg

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
)

// Push repos to new branches on target projects
func PushLatest(archives []ArchiveInfo) error {
	for _, archive := range archives {
		var stdout, stderr bytes.Buffer
		args := []string{
			"-c",
			fmt.Sprintf("%s && %s && %s",
				fmt.Sprint("git init --initial-branch=main"),
				fmt.Sprintf("git checkout -b commercial_%s", archive.ShortSHA),
				fmt.Sprintf("git remote add fedramp %s", archive.RemoteUrl),
			),
		}
		cmd := exec.Command("/bin/sh", args...)
		cmd.Dir = archive.DirPath
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err := cmd.Run()
		if err != nil {
			log.Println(fmt.Sprintf(
				"STDOUT: %s\n\nSTDERR: %s",
				stdout.String(),
				stderr.String(),
			))
			return err
		}
	}
	return nil
}
