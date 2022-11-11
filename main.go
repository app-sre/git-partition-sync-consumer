package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/app-sre/git-partition-sync-consumer/pkg"
)

func main() {
	var dryRun bool
	flag.BoolVar(&dryRun, "dry-run", false, "If true, will only print planned actions")
	flag.Parse()

	// define vars to look for and any defaults
	envVars, err := getEnvVars(map[string]string{
		"AWS_ACCESS_KEY_ID":     "",
		"AWS_SECRET_ACCESS_KEY": "",
		"AWS_REGION":            "",
		"AWS_S3_BUCKET":         "",
		"GITLAB_BASE_URL":       "",
		"GITLAB_USERNAME":       "",
		"GITLAB_TOKEN":          "",
		"PRIVATE_KEY":           "",
		"RECONCILE_SLEEP_TIME":  "5m",
		"WORKDIR":               "/working",
	})
	if err != nil {
		log.Fatalln(err)
	}

	sleepDuration, err := time.ParseDuration(envVars["RECONCILE_SLEEP_TIME"])
	if err != nil {
		log.Fatalln(err)
	}

	downloader, err := pkg.NewDownloader(
		envVars["AWS_ACCESS_KEY_ID"],
		envVars["AWS_SECRET_ACCESS_KEY"],
		envVars["AWS_REGION"],
		envVars["AWS_S3_BUCKET"],
		envVars["GITLAB_BASE_URL"],
		envVars["GITLAB_USERNAME"],
		envVars["GITLAB_TOKEN"],
		envVars["PRIVATE_KEY"],
		envVars["WORKDIR"],
	)
	if err != nil {
		log.Fatalln(err)
	}

	for {
		ctx := context.Background()
		err = downloader.Run(ctx, dryRun)
		if err != nil {
			log.Fatalln(err)
		}

		if dryRun {
			return
		} else {
			time.Sleep(sleepDuration)
		}
	}
}

// iterate through keys of desired env variables and look up values
func getEnvVars(vars map[string]string) (map[string]string, error) {
	result := make(map[string]string)
	for k := range vars {
		val := os.Getenv(k)
		if val == "" {
			// check if optional (default exists)
			if vars[k] != "" {
				result[k] = vars[k]
			} else {
				return nil, errors.New(
					fmt.Sprintf("Required environment variable missing: %s", k))
			}
		} else {
			result[k] = val
		}
	}
	return result, nil
}
