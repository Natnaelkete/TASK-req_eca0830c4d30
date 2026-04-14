package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
)

const (
	// HotRetentionDays is the number of days data stays in hot storage (fast queries).
	HotRetentionDays = 90
	// ColdRetentionYears is the number of years data stays in cold storage before purge.
	ColdRetentionYears = 3
)

// RetentionService manages monitoring data lifecycle: monthly partitioning, hot/cold retention.
type RetentionService struct {
	db *gorm.DB
}

func NewRetentionService(db *gorm.DB) *RetentionService {
	return &RetentionService{db: db}
}

// EnsurePartitions extends monthly partitions for monitoring_data for the next 3 months.
// The base partitions are created by migration 002. This worker reorganizes the pmax
// overflow partition to add new monthly boundaries as needed.
func (s *RetentionService) EnsurePartitions(ctx context.Context) error {
	now := time.Now()

	for i := 0; i < 3; i++ {
		target := now.AddDate(0, i+1, 0)
		partName := fmt.Sprintf("p%d%02d", target.Year(), target.Month())
		boundary := fmt.Sprintf("%d-%02d-01", target.Year(), target.Month())

		// Reorganize pmax to carve out a new named partition before the overflow.
		// If the partition already exists, MySQL will return an error which we ignore.
		sql := fmt.Sprintf(
			"ALTER TABLE monitoring_data REORGANIZE PARTITION pmax INTO (PARTITION %s VALUES LESS THAN (TO_DAYS('%s')), PARTITION pmax VALUES LESS THAN MAXVALUE)",
			partName, boundary,
		)
		if err := s.db.WithContext(ctx).Exec(sql).Error; err != nil {
			// Ignore "duplicate partition name" or "already exists" errors
			if !containsDuplicateMsg(err.Error()) && !containsStr(err.Error(), "already exists") {
				log.Printf("partition %s maintenance note: %v", partName, err)
			}
		}
	}
	return nil
}

func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

// ArchiveColdData marks monitoring data older than HotRetentionDays as cold
// by moving it to a separate archive table or flagging it.
// In this implementation, we use a soft archive: data older than hot threshold
// is moved to monitoring_data_archive table.
func (s *RetentionService) ArchiveColdData(ctx context.Context) (int64, error) {
	hotCutoff := time.Now().AddDate(0, 0, -HotRetentionDays)

	// Insert into archive from main table (archive table created by migration 002)
	insertSQL := `INSERT IGNORE INTO monitoring_data_archive SELECT * FROM monitoring_data WHERE event_time < ?`
	result := s.db.WithContext(ctx).Exec(insertSQL, hotCutoff)
	if result.Error != nil {
		return 0, fmt.Errorf("archive cold data: %w", result.Error)
	}
	archived := result.RowsAffected

	// Delete archived records from main table
	deleteSQL := `DELETE FROM monitoring_data WHERE event_time < ?`
	if err := s.db.WithContext(ctx).Exec(deleteSQL, hotCutoff).Error; err != nil {
		return archived, fmt.Errorf("delete archived data from hot: %w", err)
	}

	return archived, nil
}

// PurgeColdData removes archived data older than ColdRetentionYears.
func (s *RetentionService) PurgeColdData(ctx context.Context) (int64, error) {
	purgeCutoff := time.Now().AddDate(-ColdRetentionYears, 0, 0)

	result := s.db.WithContext(ctx).Exec(
		`DELETE FROM monitoring_data_archive WHERE event_time < ?`, purgeCutoff,
	)
	if result.Error != nil {
		return 0, fmt.Errorf("purge cold data: %w", result.Error)
	}
	return result.RowsAffected, nil
}

// StartRetentionWorker starts a background goroutine that runs archival and purge daily.
func (s *RetentionService) StartRetentionWorker(ctx context.Context) {
	go func() {
		// Run once at startup
		if err := s.EnsurePartitions(ctx); err != nil {
			log.Printf("initial partition setup: %v", err)
		}

		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				log.Println("retention worker stopped")
				return
			case <-ticker.C:
				// Ensure future partitions
				if err := s.EnsurePartitions(ctx); err != nil {
					log.Printf("partition maintenance error: %v", err)
				}

				// Archive hot -> cold
				archived, err := s.ArchiveColdData(ctx)
				if err != nil {
					log.Printf("archive cold data error: %v", err)
				} else if archived > 0 {
					log.Printf("archived %d monitoring records to cold storage", archived)
				}

				// Purge expired cold data
				purged, err := s.PurgeColdData(ctx)
				if err != nil {
					log.Printf("purge cold data error: %v", err)
				} else if purged > 0 {
					log.Printf("purged %d expired archive records", purged)
				}
			}
		}
	}()
}
