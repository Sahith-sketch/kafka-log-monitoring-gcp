package services

import (
	"context"
	"fmt"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"audit-logs-monitoring/pkg/logger"
)

type SecretManager struct {
	client    *secretmanager.Client
	projectID string
}

func NewSecretManager(projectID string) (*SecretManager, error) {
	ctx := context.Background()
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create secret manager client: %w", err)
	}

	return &SecretManager{
		client:    client,
		projectID: projectID,
	}, nil
}

func (sm *SecretManager) GetSecret(secretName string) (string, error) {
	if secretName == "" {
		return "", nil
	}

	ctx := context.Background()
	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: fmt.Sprintf("projects/%s/secrets/%s/versions/latest", sm.projectID, secretName),
	}

	result, err := sm.client.AccessSecretVersion(ctx, req)
	if err != nil {
		logger.Infof("Failed to access secret %s: %v", secretName, err)
		return "", fmt.Errorf("failed to access secret %s: %w", secretName, err)
	}

	return string(result.Payload.Data), nil
}

func (sm *SecretManager) Close() error {
	if err := sm.client.Close(); err != nil {
		logger.Infof("Failed to close secret manager client: %v", err)
		return fmt.Errorf("failed to close secret manager client: %w", err)
	}
	return nil
}