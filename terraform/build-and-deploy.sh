#!/bin/bash

set -e

PROJECT_ID=${1:-""}
REGION=${2:-"us-central1"}

if [ -z "$PROJECT_ID" ]; then
    echo "Usage: $0 <PROJECT_ID> [REGION]"
    echo "Example: $0 my-gcp-project us-central1"
    exit 1
fi

echo "🏗️  Building and deploying complete infrastructure with application"

# Step 1: Deploy infrastructure first
echo "📦 Step 1: Deploying infrastructure..."
./deploy.sh $PROJECT_ID $REGION

# Get repository URL from Terraform output
REPO_URL=$(terraform output -raw artifact_registry_url)
IMAGE_TAG="$REPO_URL/audit-processor:latest"

echo "🐳 Step 2: Building and pushing Docker image..."

# Configure Docker authentication
gcloud auth configure-docker $REGION-docker.pkg.dev

# Build and push image
cd ..
docker build -t $IMAGE_TAG .
docker push $IMAGE_TAG

# Step 3: Update Cloud Run with new image
echo "🚀 Step 3: Updating Cloud Run service..."
cd terraform
terraform apply -var="project_id=$PROJECT_ID" -var="cloud_run_image=$IMAGE_TAG" -auto-approve

echo "✅ Complete deployment finished!"
echo "🌐 Cloud Run URL: $(terraform output -raw cloud_run_url)"