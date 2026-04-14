-- 002_indicator_versions_and_partitioning.sql
-- Adds indicator version management tables, monitoring data monthly partitioning,
-- hot/cold archival table, and user email encryption fields.

-- ============================================================
-- 1. Indicator Definitions (analysis metric definitions with version tracking)
-- ============================================================

CREATE TABLE IF NOT EXISTS `indicator_definitions` (
    `id`          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `code`        VARCHAR(100)    NOT NULL,
    `name`        VARCHAR(255)    NOT NULL,
    `description` TEXT            NULL,
    `unit`        VARCHAR(50)     NULL,
    `formula`     TEXT            NULL,
    `category`    VARCHAR(100)    NULL,
    `status`      VARCHAR(50)     NOT NULL DEFAULT 'active',
    `created_by`  BIGINT UNSIGNED NOT NULL,
    `created_at`  DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    `updated_at`  DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    PRIMARY KEY (`id`),
    UNIQUE INDEX `idx_indicator_code` (`code`),
    INDEX `idx_indicator_category` (`category`),
    INDEX `idx_indicator_created_by` (`created_by`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================
-- 2. Indicator Versions (audit-diff chain for every indicator change)
-- ============================================================

CREATE TABLE IF NOT EXISTS `indicator_versions` (
    `id`            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `indicator_id`  BIGINT UNSIGNED NOT NULL,
    `version`       INT             NOT NULL,
    `name`          VARCHAR(255)    NOT NULL,
    `description`   TEXT            NULL,
    `unit`          VARCHAR(50)     NULL,
    `formula`       TEXT            NULL,
    `category`      VARCHAR(100)    NULL,
    `diff_summary`  TEXT            NULL,
    `modified_by`   BIGINT UNSIGNED NOT NULL,
    `modified_at`   DATETIME(3)     NOT NULL,
    `created_at`    DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    PRIMARY KEY (`id`),
    INDEX `idx_iv_indicator` (`indicator_id`),
    INDEX `idx_iv_modified_by` (`modified_by`),
    CONSTRAINT `fk_iv_indicator` FOREIGN KEY (`indicator_id`) REFERENCES `indicator_definitions` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================
-- 3. User email encryption support
-- ============================================================

ALTER TABLE `users` ADD COLUMN IF NOT EXISTS `email_hash` VARCHAR(64) NOT NULL DEFAULT '' AFTER `email`;
ALTER TABLE `users` ADD UNIQUE INDEX IF NOT EXISTS `idx_users_email_hash` (`email_hash`);

-- ============================================================
-- 4. Monitoring Data: Monthly Partitioning (deterministic bootstrap)
--
-- MySQL requires that partition columns be part of the primary key and
-- does not support foreign keys on partitioned tables.
-- This migration restructures monitoring_data for RANGE partitioning
-- on event_time with a 12-month rolling window plus overflow partition.
-- ============================================================

-- 4a. Drop foreign keys that are incompatible with partitioning
ALTER TABLE `monitoring_data` DROP FOREIGN KEY IF EXISTS `fk_monitoring_device`;
ALTER TABLE `monitoring_data` DROP FOREIGN KEY IF EXISTS `fk_monitoring_plot`;

-- 4b. Restructure primary key to include event_time (required for partitioning)
ALTER TABLE `monitoring_data` DROP PRIMARY KEY, ADD PRIMARY KEY (`id`, `event_time`);

-- 4c. Apply deterministic monthly partitioning (12 months + overflow)
ALTER TABLE `monitoring_data` PARTITION BY RANGE (TO_DAYS(`event_time`)) (
    PARTITION p202601 VALUES LESS THAN (TO_DAYS('2026-02-01')),
    PARTITION p202602 VALUES LESS THAN (TO_DAYS('2026-03-01')),
    PARTITION p202603 VALUES LESS THAN (TO_DAYS('2026-04-01')),
    PARTITION p202604 VALUES LESS THAN (TO_DAYS('2026-05-01')),
    PARTITION p202605 VALUES LESS THAN (TO_DAYS('2026-06-01')),
    PARTITION p202606 VALUES LESS THAN (TO_DAYS('2026-07-01')),
    PARTITION p202607 VALUES LESS THAN (TO_DAYS('2026-08-01')),
    PARTITION p202608 VALUES LESS THAN (TO_DAYS('2026-09-01')),
    PARTITION p202609 VALUES LESS THAN (TO_DAYS('2026-10-01')),
    PARTITION p202610 VALUES LESS THAN (TO_DAYS('2026-11-01')),
    PARTITION p202611 VALUES LESS THAN (TO_DAYS('2026-12-01')),
    PARTITION p202612 VALUES LESS THAN (TO_DAYS('2027-01-01')),
    PARTITION pmax    VALUES LESS THAN MAXVALUE
);

-- ============================================================
-- 5. Monitoring Data Archive (cold storage for data older than 90 days)
--    Structurally identical to monitoring_data but not partitioned.
-- ============================================================

CREATE TABLE IF NOT EXISTS `monitoring_data_archive` (
    `id`          BIGINT UNSIGNED NOT NULL,
    `source_id`   VARCHAR(255)    NOT NULL,
    `device_id`   BIGINT UNSIGNED NOT NULL,
    `plot_id`     BIGINT UNSIGNED NOT NULL,
    `metric_code` VARCHAR(100)    NOT NULL,
    `value`       DOUBLE          NOT NULL,
    `unit`        VARCHAR(50)     NULL,
    `event_time`  DATETIME(3)     NOT NULL,
    `tags`        TEXT            NULL,
    `created_at`  DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    PRIMARY KEY (`id`, `event_time`),
    INDEX `idx_archive_core` (`device_id`, `plot_id`, `metric_code`, `event_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
