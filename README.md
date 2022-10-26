# github.com/app-sre/gitlab-sync-s3-pull
Utility for pulling git archives from s3 and pushing to remote. Reliant on S3 object key format outputted by [gitlab-sync-s3-push](https://github.com/dwelch0/gitlab-sync-s3-push)

## Environment Variables
* AWS_ACCESS_KEY_ID (needs s3 read permission)
* AWS_SECRET_ACCESS_KEY
* AWS_REGION
* AWS_S3_BUCKET
* GITLAB_BASE_URL
* GITLAB_USERNAME
* GITLAB_TOKEN (needs repo write permission)
* PRIVATE_KEY
* RECONCILE_SLEEP_TIME
* WORKDIR

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
    quay.io/app-sre/gitlab-sync-pull:latest -dry-run
```