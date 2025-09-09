#!/bin/bash

set -e

PROJECT_ID=${1:-""}

if [ -z "$PROJECT_ID" ]; then
    echo "Usage: $0 <PROJECT_ID>"
    echo "Example: $0 my-gcp-project"
    exit 1
fi

echo "🗑️  Destroying GCP Audit Log Processor Infrastructure"
echo "Project ID: $PROJECT_ID"

read -p "Are you sure you want to destroy all infrastructure? (yes/no): " confirm
if [ "$confirm" != "yes" ]; then
    echo "Aborted."
    exit 1
fi

# Initialize Terraform
echo "🔧 Initializing Terraform..."
terraform init -backend-config="bucket=terraform-state-$PROJECT_ID"

# Plan destroy
echo "📋 Planning destruction..."
terraform plan -destroy -var="project_id=$PROJECT_ID"

# Apply destroy
echo "💥 Destroying infrastructure..."
terraform destroy -var="project_id=$PROJECT_ID" -auto-approve

echo "✅ Infrastructure destruction completed!"