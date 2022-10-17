package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/app-sre/gitlab-sync-pull/pkg"
)

const (
	AWS_S3_BUCKET          = "AWS_S3_BUCKET"
	PRIVATE_GPG_PATH       = "PRIVATE_GPG_PATH"
	PRIVATE_GPG_PASSPHRASE = "PRIVATE_GPG_PASSPHRASE"
	RECONCILE_SLEEP_TIME   = "RECONCILE_SLEEP_TIME"
	GITLAB_USERNAME        = "GITLAB_USERNAME"
	GITLAB_TOKEN           = "GITLAB_TOKEN"
)

func main() {
	var dryRun bool
	flag.BoolVar(&dryRun, "dry-run", false, "If true, will only print planned actions")
	flag.Parse()

	// get necessary env variables
	path, _ := os.LookupEnv(PRIVATE_GPG_PATH)
	if path == "" {
		log.Fatalf("Missing environment variable: %s", PRIVATE_GPG_PATH)
	}
	passphrase, _ := os.LookupEnv(PRIVATE_GPG_PASSPHRASE)
	if passphrase == "" {
		log.Fatalf("Missing environment variable: %s", PRIVATE_GPG_PASSPHRASE)
	}
	bucket, _ := os.LookupEnv(AWS_S3_BUCKET)
	if bucket == "" {
		log.Fatalf("Missing environment variable: %s", AWS_S3_BUCKET)
	}
	gitlabUsername, _ := os.LookupEnv(GITLAB_USERNAME)
	if gitlabUsername == "" {
		log.Fatalf("Missing environment variable: %s", GITLAB_USERNAME)
	}
	gitlabToken, _ := os.LookupEnv(GITLAB_TOKEN)
	if gitlabUsername == "" {
		log.Fatalf("Missing environment variable: %s", GITLAB_TOKEN)
	}
	sleep, _ := os.LookupEnv(RECONCILE_SLEEP_TIME)
	if sleep == "" {
		log.Fatalf("Missing environment variable: %s", RECONCILE_SLEEP_TIME)
	}
	sleepDuration, err := time.ParseDuration(sleep)
	if err != nil {
		log.Fatalln(err)
	}

	log.Fatal(run(bucket, path, passphrase, gitlabUsername, gitlabToken, dryRun, sleepDuration))
}

func run(bucket, gpgPath, gpgPassphrase, gitlabUsername, gitlabToken string, dryRun bool, sleep time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()
	s3, err := pkg.NewS3Helper(ctx, bucket)
	if err != nil {
		return err
	}

	first := true
	for {
		// janky if-else so `continue`s can be utilized w/ skipping sleep
		if first {
			first = false
		} else {
			time.Sleep(sleep)
		}
		log.Println("Beginning sync...")

		encryptedUpdates, err := s3.GetUpdatedObjects(ctx)
		if err != nil {
			log.Println(err)
			continue
		}
		if len(encryptedUpdates) < 1 {
			continue
		}

		gpg, err := pkg.NewGpgHelper(gpgPath, gpgPassphrase)
		if err != nil {
			log.Println(err)
			continue
		}

		decryptedObjs, err := gpg.DecryptBundles(encryptedUpdates)
		if err != nil {
			log.Println(err)
			continue
		}

		archives, err := pkg.Extract(decryptedObjs)
		if err != nil {
			log.Println(err)
			continue
		}
		if dryRun {
			// print to be pushed repos

			return nil
		}

		err = pkg.PushLatest(gitlabUsername, gitlabToken, archives)
		if err != nil {
			log.Println(err)
			continue
		}

		s3.UpdateCache()

		log.Println("Sync completed")
		for _, archive := range archives {
			log.Println(
				fmt.Sprintf("New branch pushed for %s with short sha %s",
					archive.RemoteSshUri,
					archive.ShortSHA),
			)
		}
	}
}
