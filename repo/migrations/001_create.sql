-- 001_create.sql
-- Creates all tables for the Agricultural Research Data Platform.

CREATE TABLE IF NOT EXISTS `users` (
    `id`            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `username`      VARCHAR(100)    NOT NULL,
    `email`         VARCHAR(500)    NOT NULL,
    `email_hash`    VARCHAR(64)     NOT NULL DEFAULT '',
    `password_hash` VARCHAR(255)    NOT NULL,
    `role`          VARCHAR(50)     NOT NULL DEFAULT 'researcher',
    `created_at`    DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    `updated_at`    DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    PRIMARY KEY (`id`),
    UNIQUE INDEX `idx_users_username` (`username`),
    UNIQUE INDEX `idx_users_email_hash` (`email_hash`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `plots` (
    `id`         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `name`       VARCHAR(200)    NOT NULL,
    `location`   VARCHAR(255)    NULL,
    `area`       DOUBLE          NULL DEFAULT 0,
    `soil_type`  VARCHAR(100)    NULL,
    `crop_type`  VARCHAR(100)    NULL,
    `user_id`    BIGINT UNSIGNED NOT NULL,
    `created_at` DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    `updated_at` DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    PRIMARY KEY (`id`),
    INDEX `idx_plots_user_id` (`user_id`),
    CONSTRAINT `fk_plots_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `devices` (
    `id`            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `name`          VARCHAR(200)    NOT NULL,
    `type`          VARCHAR(100)    NOT NULL,
    `serial_number` VARCHAR(100)    NOT NULL,
    `plot_id`       BIGINT UNSIGNED NOT NULL,
    `status`        VARCHAR(50)     NOT NULL DEFAULT 'active',
    `installed_at`  DATETIME(3)     NULL,
    `created_at`    DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    `updated_at`    DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    PRIMARY KEY (`id`),
    UNIQUE INDEX `idx_devices_serial` (`serial_number`),
    INDEX `idx_devices_plot_id` (`plot_id`),
    CONSTRAINT `fk_devices_plot` FOREIGN KEY (`plot_id`) REFERENCES `plots` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `metrics` (
    `id`          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `device_id`   BIGINT UNSIGNED NOT NULL,
    `metric_type` VARCHAR(100)    NOT NULL,
    `value`       DOUBLE          NOT NULL,
    `unit`        VARCHAR(50)     NULL,
    `event_time`  DATETIME(3)     NOT NULL,
    `created_at`  DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    PRIMARY KEY (`id`),
    INDEX `idx_metrics_device_id` (`device_id`),
    INDEX `idx_metrics_type` (`metric_type`),
    INDEX `idx_metrics_event_time` (`event_time`),
    CONSTRAINT `fk_metrics_device` FOREIGN KEY (`device_id`) REFERENCES `devices` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `tasks` (
    `id`           BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `title`        VARCHAR(255)    NOT NULL,
    `description`  TEXT            NULL,
    `object_id`    BIGINT UNSIGNED NOT NULL,
    `object_type`  VARCHAR(100)    NOT NULL,
    `cycle_type`   VARCHAR(50)     NULL,
    `status`       VARCHAR(50)     NOT NULL DEFAULT 'pending',
    `assigned_to`  BIGINT UNSIGNED NULL DEFAULT 0,
    `reviewer_id`  BIGINT UNSIGNED NULL,
    `due_start`    DATETIME(3)     NULL,
    `due_end`      DATETIME(3)     NULL,
    `submitted_at` DATETIME(3)     NULL,
    `overdue_at`   DATETIME(3)     NULL,
    `created_at`   DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    `updated_at`   DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    PRIMARY KEY (`id`),
    INDEX `idx_tasks_status` (`status`),
    INDEX `idx_tasks_assigned` (`assigned_to`),
    INDEX `idx_tasks_object` (`object_id`),
    INDEX `idx_tasks_reviewer` (`reviewer_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `audit_logs` (
    `id`          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `user_id`     BIGINT UNSIGNED NULL DEFAULT 0,
    `action`      VARCHAR(100)    NOT NULL,
    `resource`    VARCHAR(100)    NOT NULL,
    `resource_id` BIGINT UNSIGNED NULL DEFAULT 0,
    `ip_address`  VARCHAR(45)     NULL,
    `created_at`  DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    PRIMARY KEY (`id`),
    INDEX `idx_audit_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `alerts` (
    `id`          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `device_id`   BIGINT UNSIGNED NOT NULL,
    `metric_type` VARCHAR(100)    NOT NULL,
    `value`       DOUBLE          NOT NULL,
    `threshold`   DOUBLE          NOT NULL,
    `level`       VARCHAR(50)     NOT NULL DEFAULT 'warning',
    `message`     VARCHAR(500)    NULL,
    `resolved`    TINYINT(1)      NOT NULL DEFAULT 0,
    `created_at`  DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    PRIMARY KEY (`id`),
    INDEX `idx_alerts_device_id` (`device_id`),
    INDEX `idx_alerts_level` (`level`),
    CONSTRAINT `fk_alerts_device` FOREIGN KEY (`device_id`) REFERENCES `devices` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `messages` (
    `id`          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `sender_id`   BIGINT UNSIGNED NOT NULL,
    `receiver_id` BIGINT UNSIGNED NOT NULL,
    `plot_id`     BIGINT UNSIGNED NULL,
    `content`     TEXT            NOT NULL,
    `read`        TINYINT(1)      NOT NULL DEFAULT 0,
    `created_at`  DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    PRIMARY KEY (`id`),
    INDEX `idx_messages_sender` (`sender_id`),
    INDEX `idx_messages_receiver` (`receiver_id`),
    INDEX `idx_messages_plot` (`plot_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `monitoring_data` (
    `id`          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `source_id`   VARCHAR(255)    NOT NULL,
    `device_id`   BIGINT UNSIGNED NOT NULL,
    `plot_id`     BIGINT UNSIGNED NOT NULL,
    `metric_code` VARCHAR(100)    NOT NULL,
    `value`       DOUBLE          NOT NULL,
    `unit`        VARCHAR(50)     NULL,
    `event_time`  DATETIME(3)     NOT NULL,
    `tags`        TEXT            NULL,
    `created_at`  DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    PRIMARY KEY (`id`),
    UNIQUE INDEX `idx_monitoring_idempotent` (`source_id`, `metric_code`, `event_time`),
    INDEX `idx_monitoring_core` (`device_id`, `plot_id`, `metric_code`, `event_time`),
    CONSTRAINT `fk_monitoring_device` FOREIGN KEY (`device_id`) REFERENCES `devices` (`id`) ON DELETE CASCADE,
    CONSTRAINT `fk_monitoring_plot`   FOREIGN KEY (`plot_id`)   REFERENCES `plots` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `dashboard_configs` (
    `id`         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `user_id`    BIGINT UNSIGNED NOT NULL,
    `name`       VARCHAR(255)    NOT NULL,
    `config`     TEXT            NOT NULL,
    `created_at` DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    `updated_at` DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    PRIMARY KEY (`id`),
    INDEX `idx_dashboard_user` (`user_id`),
    CONSTRAINT `fk_dashboard_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `results` (
    `id`                  BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `type`                VARCHAR(50)     NOT NULL,
    `plot_id`             BIGINT UNSIGNED NOT NULL,
    `task_id`             BIGINT UNSIGNED NULL,
    `title`               VARCHAR(255)    NOT NULL,
    `summary`             TEXT            NULL,
    `fields`              TEXT            NULL,
    `status`              VARCHAR(50)     NOT NULL DEFAULT 'draft',
    `submitter_id`        BIGINT UNSIGNED NOT NULL,
    `notes`               TEXT            NULL,
    `archived_at`         DATETIME(3)     NULL,
    `invalidated_reason`  TEXT            NULL,
    `invalidated_by`      BIGINT UNSIGNED NULL,
    `invalidated_at`      DATETIME(3)     NULL,
    `created_by`          BIGINT UNSIGNED NOT NULL,
    `created_at`          DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    `updated_at`          DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    PRIMARY KEY (`id`),
    INDEX `idx_results_plot` (`plot_id`),
    INDEX `idx_results_task` (`task_id`),
    INDEX `idx_results_status` (`status`),
    INDEX `idx_results_submitter` (`submitter_id`),
    INDEX `idx_results_created_by` (`created_by`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
