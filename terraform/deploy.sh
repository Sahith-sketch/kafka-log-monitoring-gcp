#!/bin/bash

set -e

PROJECT_ID=${1:-""}
REGION=${2:-"us-central1"}

if [ -z "$PROJECT_ID" ]; then
    echo "Usage: $0 <PROJECT_ID> [REGION]"
    echo "Example: $0 my-gcp-project us-central1"
    exit 1
fi

echo "🚀 Deploying GCP Audit Log Processor Infrastructure"
echo "Project ID: $PROJECT_ID"
echo "Region: $REGION"

# Create terraform.tfvars if it doesn't exist
if [ ! -f terraform.tfvars ]; then
    echo "📝 Creating terraform.tfvars from example..."
    cp terraform.tfvars.example terraform.tfvars
    sed -i "s/your-gcp-project-id/$PROJECT_ID/g" terraform.tfvars
    sed -i "s/us-central1/$REGION/g" terraform.tfvars
    echo "⚠️  Please update terraform.tfvars with your specific values before proceeding"
    exit 1
fi

# Create state bucket if it doesn't exist
echo "📦 Creating Terraform state bucket..."
gsutil mb -p $PROJECT_ID gs://terraform-state-$PROJECT_ID 2>/dev/null || echo "Bucket already exists"
gsutil versioning set on gs://terraform-state-$PROJECT_ID

# Initialize Terraform
echo "🔧 Initializing Terraform..."
terraform init -reconfigure -backend-config="bucket=terraform-state-$PROJECT_ID" -backend-config="prefix=audit-processor"

# Plan deployment
echo "📋 Planning deployment..."
terraform plan -var="project_id=$PROJECT_ID"

# Apply deployment
echo "🚀 Applying deployment..."
terraform apply -var="project_id=$PROJECT_ID" -auto-approve

echo "✅ Infrastructure deployment completed!"
echo "📊 Outputs:"
terraform output