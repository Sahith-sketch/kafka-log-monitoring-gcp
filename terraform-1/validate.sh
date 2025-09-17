#!/bin/bash

echo "🔍 Validating Terraform Configuration..."

# Check required files
echo "📁 Checking required files..."
required_files=("backend.tf" "providers.tf" "variables.tf" "outputs.tf")
for file in "${required_files[@]}"; do
    if [ -f "$file" ]; then
        echo "✅ $file"
    else
        echo "❌ Missing: $file"
        exit 1
    fi
done

# Check cert files
echo "📜 Checking certificate files..."
cert_files=("../certs/ca.crt" "../certs/cert.pem" "../certs/cert.key")
for cert in "${cert_files[@]}"; do
    if [ -f "$cert" ]; then
        echo "✅ $cert"
    else
        echo "❌ Missing: $cert"
        exit 1
    fi
done

# Terraform validation
echo "🔧 Running Terraform validation..."
terraform init -backend=false
terraform validate

if [ $? -eq 0 ]; then
    echo "✅ Terraform configuration is valid!"
else
    echo "❌ Terraform validation failed!"
    exit 1
fi

echo "🎉 All validations passed!"