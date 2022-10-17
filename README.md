# gitlab-sync-pull
Utility for pulling git archives from s3 and pushing to remote

## Environment Variables
* AWS_ACCESS_KEY_ID (needs s3 read permission)
* AWS_SECRET_ACCESS_KEY
* AWS_REGION
* AWS_S3_BUCKET
* PRIVATE_GPG_PATH (key should be armored)
* PRIVATE_GPG_PASSPHRASE
* GITLAB_USERNAME
* GITLAB_TOKEN (needs repo write permission)
* RECONCILE_SLEEP_TIME

## Execute
```
podman run -t \
    -v "$PRIVATE_GPG_PATH":/"$PRIVATE_GPG_PATH" \
    -e AWS_ACCESS_KEY_ID="$AWS_ACCESS_KEY_ID" \
    -e AWS_SECRET_ACCESS_KEY="$AWS_SECRET_ACCESS_KEY" \
    -e AWS_REGION="$AWS_REGION" \
    -e AWS_S3_BUCKET="$AWS_S3_BUCKET" \
    -e PRIVATE_GPG_PATH="$PRIVATE_GPG_PATH" \
    -e PRIVATE_GPG_PASSPHRASE="$PRIVATE_GPG_PASSPHRASE" \
    -e GITLAB_USERNAME="$GITLAB_USERNAME" \
    -e GITLAB_TOKEN="$GITLAB_TOKEN" \
    -e RECONCILE_SLEEP_TIME="$RECONCILE_SLEEP_TIME" \
    quay.io/app-sre/gitlab-sync-pull:latest -dry-run
```