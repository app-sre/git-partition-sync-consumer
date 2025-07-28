# Git Partition Sync - Consumer
Utility for pulling git archives from s3 and pushing to remote. Reliant on S3 object key format outputted by [gitlab-sync-s3-push](https://github.com/app-sre/git-partition-sync-producer)

![gitlab-sync-diagram](gitsync-diagram.png)

## Environment Variables

### Required
* AWS_ACCESS_KEY_ID (needs s3 read permission)
* AWS_SECRET_ACCESS_KEY
* AWS_REGION
* AWS_S3_BUCKET
* GITLAB_BASE_URL
* GITLAB_USERNAME
* GITLAB_TOKEN (needs repo write permission)
* PRIVATE_KEY

### Optional
* RECONCILE_SLEEP_TIME - time between runs. defaults to 5 minutes (5m)
* WORKDIR - local directory where io operations will be performed. defaults to `/working`

## Execute
```
podman run -t \
    -e AWS_ACCESS_KEY_ID="$AWS_ACCESS_KEY_ID" \
    -e AWS_SECRET_ACCESS_KEY="$AWS_SECRET_ACCESS_KEY" \
    -e AWS_REGION="$AWS_REGION" \
    -e AWS_S3_BUCKET="$AWS_S3_BUCKET" \
    -e GITLAB_BASE_URL="$GITLAB_BASE_URL" \
    -e GITLAB_USERNAME="$GITLAB_USERNAME" \
    -e GITLAB_TOKEN="$GITLAB_TOKEN" \
    -e PRIVATE_KEY="$PRIVATE_KEY" \
    quay.io/app-sre/git-partition-sync-consumer:latest -dry-run
```
