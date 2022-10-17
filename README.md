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
