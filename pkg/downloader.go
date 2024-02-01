package pkg

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/app-sre/git-partition-sync-consumer/pkg/metrics"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Downloader struct {
	awsAccessKey string
	awsSecretKey string
	awsRegion    string
	bucket       string
	glBaseURL    string
	glUsername   string
	glToken      string
	privateKey   string
	shard        string
	workdir      string

	s3Client *s3.Client

	cache map[string]time.Time
	tmp   map[string]time.Time
}

func NewDownloader(
	awsAccessKey,
	awsSecretKey,
	awsRegion,
	bucket,
	glURL,
	glUsername,
	glToken,
	metricsPort,
	privateKey,
	shard,
	workdir string,
	runOnce bool) (*Downloader, error) {

	err := os.Mkdir(workdir, 0755)
	if err != nil {
		return nil, err
	}

	if !runOnce {
		metrics.Start(metricsPort)
	}

	return &Downloader{
		awsRegion:    awsRegion,
		awsAccessKey: awsAccessKey,
		awsSecretKey: awsSecretKey,
		bucket:       bucket,
		glBaseURL:    glURL,
		glUsername:   glUsername,
		glToken:      glToken,
		privateKey:   privateKey,
		workdir:      workdir,
		cache:        make(map[string]time.Time),
		tmp:          make(map[string]time.Time),
	}, nil
}

func (d *Downloader) Run(ctx context.Context, dryRun, runOnce bool) error {
	log.Println("Beginning sync...")
	start := time.Now()

	defer d.clear()

	status := 1
	if !runOnce {
		defer func() {
			metrics.RecordMetrics(d.shard, status, time.Since(start))
		}()
	}

	d.initS3Client()

	ctxCancel, cancel := context.WithCancel(ctx)
	defer cancel()

	encryptedUpdates, err := d.getUpdatedObjects(ctxCancel)
	if err != nil {
		return err
	}
	// nothing to do
	if len(encryptedUpdates) < 1 {
		log.Println("Everything is up to date. Exiting early.")
		status = 0
		return nil
	}

	for _, obj := range encryptedUpdates {
		defer obj.Reader().Close()
	}

	decryptedObjs, err := d.decryptBundles(encryptedUpdates)
	if err != nil {
		return err
	}

	archives, err := d.extract(decryptedObjs)
	if err != nil {
		return err
	}

	if dryRun {
		for _, archive := range archives {
			log.Println(
				fmt.Sprintf("[DRY-RUN] pushed to %s on branch %s with short sha %s",
					fmt.Sprintf("%s/%s/%s", d.glBaseURL, archive.RemoteGroup, archive.RemoteName),
					archive.RemoteBranch,
					archive.ShortSHA),
			)
		}
		status = 0
		return nil
	}

	err = d.pushLatest(archives)
	if err != nil {
		return err
	}

	d.updateCache()

	for _, archive := range archives {
		log.Println(
			fmt.Sprintf("successfully pushed latest to %s on branch %s with short sha %s",
				fmt.Sprintf("%s/%s/%s", d.glBaseURL, archive.RemoteGroup, archive.RemoteName),
				archive.RemoteBranch,
				archive.ShortSHA),
		)
	}

	log.Println("Sync successfully completed.")
	status = 0

	return nil
}

// clean target working directory
func (d *Downloader) clean(directory string) error {
	err := os.RemoveAll(directory)
	if err != nil {
		return err
	}

	err = os.Mkdir(directory, 0775)
	if err != nil {
		return err
	}
	return nil
}

// clear all items in working directory
func (d *Downloader) clear() error {
	err := os.RemoveAll(UNTAR_DIRECTORY)
	if err != nil {
		return err
	}
	return nil
}
