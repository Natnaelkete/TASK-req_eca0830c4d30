-- 001_create.sql
-- Creates all tables for the Agricultural Research Data Platform.

CREATE TABLE IF NOT EXISTS `users` (
    `id`            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `username`      VARCHAR(100)    NOT NULL,
    `email`         VARCHAR(255)    NOT NULL,
    `password_hash` VARCHAR(255)    NOT NULL,
    `role`          VARCHAR(50)     NOT NULL DEFAULT 'researcher',
    `created_at`    DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    `updated_at`    DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    PRIMARY KEY (`id`),
    UNIQUE INDEX `idx_users_username` (`username`),
    UNIQUE INDEX `idx_users_email` (`email`)
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
    `id`          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `title`       VARCHAR(255)    NOT NULL,
    `description` TEXT            NULL,
    `status`      VARCHAR(50)     NOT NULL DEFAULT 'pending',
    `assigned_to` BIGINT UNSIGNED NULL DEFAULT 0,
    `due_date`    DATETIME(3)     NULL,
    `created_at`  DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    `updated_at`  DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    PRIMARY KEY (`id`),
    INDEX `idx_tasks_status` (`status`),
    INDEX `idx_tasks_assigned` (`assigned_to`)
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
