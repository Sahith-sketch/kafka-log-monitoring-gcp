project_id        = "test-project-123"
region            = "us-central1"
environment       = "dev"
app_name          = "audit-processor"
pubsub_topic_name = "audit-logs"
cloud_run_image   = "us-central1-docker.pkg.dev/test-project-123/audit-processor-repo/audit-processor:latest"
batch_size        = 100
worker_count      = 10
log_filter        = "protoPayload.authenticationInfo.authoritySelector=\"kaiju-admin\" AND severity=\"DEFAULT\""