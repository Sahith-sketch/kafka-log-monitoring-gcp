package services

import (
	"audit-logs-monitoring/internal/models"
	"audit-logs-monitoring/pkg/logger"
	"context"
)

var targetPermissions = []string{
	"io.k8s.core.v1.secrets.list",
	"com.optum.pephccdp.v1alpha1.kafkaacls.get",
	"com.optum.pephccdp.v1alpha1.kafkaacls.create",
	"com.optum.pephccdp.v1alpha1.kafkaacls.delete",
	"com.optum.pephccdp.v1alpha1.kafkaacls.update",
}

type Processor struct {
	gcsStorage    *GCSStorage
	kafkaProducer *KafkaProducer
}

func NewProcessor(gcs *GCSStorage, kafka *KafkaProducer) *Processor {
	return &Processor{
		gcsStorage:    gcs,
		kafkaProducer: kafka,
	}
}

func (p *Processor) shouldSendToKafka(log models.AuditLog) bool {
	if log.ProtoPayload == nil || len(log.ProtoPayload.AuthorizationInfo) == 0 {
		return false
	}

	for _, authInfo := range log.ProtoPayload.AuthorizationInfo {
		for _, targetPerm := range targetPermissions {
			if authInfo.Permission == targetPerm {
				return true
			}
		}
	}

	return false
}

func (p *Processor) convertToKafkaMessage(log models.AuditLog) models.KafkaMessage {
	var projectId, clusterName, location, namespace string
	if log.Resource.Labels != nil {
		projectId = log.Resource.Labels["project_id"]
		clusterName = log.Resource.Labels["cluster_name"]
		location = log.Resource.Labels["location"]
		namespace = log.Resource.Labels["namespace"]
	}

	var principalEmail, callerIp, userAgent, serviceName, methodName, resourceName string
	var permission string
	var granted bool
	var responseStatus string = "SUCCESS"

	if log.ProtoPayload != nil {
		serviceName = log.ProtoPayload.ServiceName
		methodName = log.ProtoPayload.MethodName
		resourceName = log.ProtoPayload.ResourceName
		callerIp = log.ProtoPayload.CallerIp
		userAgent = log.ProtoPayload.CallerSuppliedUserAgent

		if log.ProtoPayload.AuthenticationInfo != nil {
			principalEmail = log.ProtoPayload.AuthenticationInfo.PrincipalEmail
		}

		for _, authInfo := range log.ProtoPayload.AuthorizationInfo {
			for _, targetPerm := range targetPermissions {
				if authInfo.Permission == targetPerm {
					permission = authInfo.Permission
					granted = authInfo.Granted
					break
				}
			}
			if permission != "" {
				break
			}
		}

		if log.ProtoPayload.Status != nil && log.ProtoPayload.Status.Code != 0 {
			responseStatus = "FAILED"
		}
	}

	return models.KafkaMessage{
		Timestamp:         log.Timestamp.Format("2006-01-02T15:04:05.000Z"),
		ProjectId:         projectId,
		ClusterName:       clusterName,
		Location:          location,
		Namespace:         namespace,
		PrincipalEmail:    principalEmail,
		CallerIp:          callerIp,
		UserAgent:         userAgent,
		ServiceName:       serviceName,
		MethodName:        methodName,
		ResourceName:      resourceName,
		Permission:        permission,
		PermissionGranted: granted,
		Severity:          log.Severity,
		ResponseStatus:    responseStatus,
		InsertId:          log.InsertId,
		LogType:           log.LogName,
	}
}

func (p *Processor) ProcessLog(ctx context.Context, log models.AuditLog) error {
	logger.WithField("insertId", log.InsertId).Info("Processing audit log")

	if ctx.Err() != nil {
		return ctx.Err()
	}

	if p.gcsStorage != nil {
		if err := p.gcsStorage.SaveLog(log); err != nil {
			logger.WithError(err).Error("Failed to save to GCS")
			return err
		}
	}

	if p.kafkaProducer != nil && p.shouldSendToKafka(log) {
		kafkaMsg := p.convertToKafkaMessage(log)
		if err := p.kafkaProducer.PublishKafkaMessage(kafkaMsg); err != nil {
			logger.WithError(err).Error("Failed to publish to Kafka")
			return err
		}
	}

	logger.WithField("insertId", log.InsertId).Info("Successfully processed audit log")
	return nil
}

func (p *Processor) Close() error {
	if p.gcsStorage != nil {
		p.gcsStorage.Close()
	}
	if p.kafkaProducer != nil {
		p.kafkaProducer.Close()
	}
	return nil
}
