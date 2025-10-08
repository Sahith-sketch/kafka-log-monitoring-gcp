package services

import (
	"audit-logs-monitoring/internal/models"
	"audit-logs-monitoring/pkg/logger"
	"context"
	"sync"
	"time"
)

type KafkaPublisher interface {
	PublishBatch(logs []models.AuditLog) error
	Close() error
}

type AlertNotifier interface {
	SendAlert(alert *models.IntrusionAlert) error
}

type StorageService interface {
	SaveBatchAsCSV(logs []models.AuditLog) error
	Close() error
}

type BatchProcessor struct {
	batchSize      int
	batchTime      time.Duration
	workerCount    int
	kafkaProducer  KafkaPublisher
	detector       *IntrusionDetector
	emailNotifier  AlertNotifier
	storageService StorageService
	logChan        chan models.AuditLog
	batchChan      chan []models.AuditLog
	wg             sync.WaitGroup
	isKafkaEnabled bool
	isGCSEnabled   bool
	isNotificationsEnabled bool
}

func NewBatchProcessor(batchSize int, batchTime time.Duration, workerCount int, kafka KafkaPublisher, detector *IntrusionDetector, notifier AlertNotifier, storage StorageService, isKafkaEnabled, isGCSEnabled, isNotificationsEnabled bool) *BatchProcessor {
	return &BatchProcessor{
		batchSize:      batchSize,
		batchTime:      batchTime,
		workerCount:    workerCount,
		kafkaProducer:  kafka,
		detector:       detector,
		emailNotifier:  notifier,
		storageService: storage,
		logChan:        make(chan models.AuditLog, batchSize*2),
		batchChan:      make(chan []models.AuditLog, workerCount),
		isKafkaEnabled: isKafkaEnabled,
		isGCSEnabled:   isGCSEnabled,
		isNotificationsEnabled: isNotificationsEnabled,
	}
}

func (bp *BatchProcessor) Start(ctx context.Context) {
	// Start batch collector
	go bp.collectBatches(ctx)

	// Start worker pool
	for i := 0; i < bp.workerCount; i++ {
		bp.wg.Add(1)
		go bp.worker(ctx, i)
	}
}

func (bp *BatchProcessor) ProcessLog(log models.AuditLog) {
	select {
	case bp.logChan <- log:
	default:
		logger.Info("Log channel full, dropping message")
	}
}

func (bp *BatchProcessor) collectBatches(ctx context.Context) {
	batch := make([]models.AuditLog, 0, bp.batchSize)
	ticker := time.NewTicker(bp.batchTime)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			if len(batch) > 0 {
				bp.batchChan <- batch
			}
			close(bp.batchChan)
			return
		case auditLog := <-bp.logChan:
			batch = append(batch, auditLog)
			if len(batch) >= bp.batchSize {
				bp.batchChan <- batch
				batch = make([]models.AuditLog, 0, bp.batchSize)
			}
		case <-ticker.C:
			if len(batch) > 0 {
				bp.batchChan <- batch
				batch = make([]models.AuditLog, 0, bp.batchSize)
			}
		}
	}
}

func (bp *BatchProcessor) worker(ctx context.Context, id int) {
	defer bp.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case batch, ok := <-bp.batchChan:
			if !ok {
				return
			}
			bp.processBatch(batch, id)
		}
	}
}

func (bp *BatchProcessor) processBatch(batch []models.AuditLog, workerID int) {
	logger.Infof("Worker %d: Starting batch processing for %d logs", workerID, len(batch))

	// Publish to Kafka if enabled
	if bp.isKafkaEnabled && bp.kafkaProducer != nil {
		if err := bp.kafkaProducer.PublishBatch(batch); err != nil {
			logger.Infof("Worker %d: Error publishing to Kafka: %v", workerID, err)
		}
	}

	// Save to GCS if enabled
	if bp.isGCSEnabled && bp.storageService != nil {
		if err := bp.storageService.SaveBatchAsCSV(batch); err != nil {
			logger.Infof("Worker %d: Error saving batch to GCS: %v", workerID, err)
		}
	}

	// Check for intrusions and send notifications if enabled
	if bp.isNotificationsEnabled && bp.emailNotifier != nil {
		for _, auditLog := range batch {
			if alert := bp.detector.DetectIntrusion(&auditLog); alert != nil {
				go func(a *models.IntrusionAlert) {
					if err := bp.emailNotifier.SendAlert(a); err != nil {
						logger.Infof("Error sending email alert: %v", err)
					}
				}(alert)
			}
		}
	}

	logger.Infof("Worker %d: Completed batch processing for %d logs", workerID, len(batch))
}

func (bp *BatchProcessor) Stop() {
	close(bp.logChan)
	bp.wg.Wait()
}
