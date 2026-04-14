# Agricultural Research Data & Results Operation Platform — API Specification

This document outlines the main external-facing API endpoints of the platform. It follows the domains described in the original prompt: Monitoring, Analysis, Communication, Results, Evaluation Tasks, Account, and Audit/Capacity.

> NOTE: This is a high-level REST specification based on the original requirements. Field names and exact payloads should be implemented consistently with the design and are not allowed to add new business features.

## 1. Common

- Base URL: `/api/v1`
- Authentication:
  - Local username + password.
  - Authenticated endpoints require an Authorization header (e.g., token or session mechanism chosen consistently in implementation).
- Error Response Format (example):
  ```json
  {
    "code": "ERROR_CODE",
    "message": "Human-readable error message"
  }
  ```

---

## 2. Account System

### 2.1 Register User

- **POST** `/api/v1/auth/register`
- **Description:** Create a new local account with username and password.
- **Request Body (JSON):**
  ```json
  {
    "username": "string",
    "password": "string"
  }
  ```
- **Validation:**
  - Password must be at least 8 characters.
  - Password must contain letters and numbers.
- **Response 201 (JSON):**
  ```json
  {
    "id": "user_id",
    "username": "string"
  }
  ```

### 2.2 Login

- **POST** `/api/v1/auth/login`
- **Description:** Authenticate with username and password.
- **Request Body (JSON):**
  ```json
  {
    "username": "string",
    "password": "string"
  }
  ```
- **Response 200 (JSON):**
  ```json
  {
    "token": "string_or_session_representation"
  }
  ```

---

## 3. Monitoring Domain

### 3.1 Batch Ingest Monitoring Data

- **POST** `/api/v1/monitoring/data/batch`
- **Description:** Asynchronously ingest monitoring data using a queue and idempotent keys.
- **Request Body (JSON):**
  ```json
  {
    "items": [
      {
        "device_id": "string_or_id",
        "plot_id": "string_or_id",
        "metric_code": "string",
        "source_id": "string",
        "event_time": "ISO8601 timestamp",
        "value": 123.45,
        "tags": { "key": "value" }
      }
    ]
  }
  ```
- **Behaviour:**
  - Items are queued for asynchronous batch writing.
  - Duplicate entries (matching `source_id`, `event_time`, `metric_code`) are not inserted twice.

### 3.2 Query Monitoring Data

- **GET** `/api/v1/monitoring/data`
- **Description:** Query monitoring data with filtering and aggregation.
- **Query Parameters (examples):**
  - `plot_id`
  - `device_id`
  - `metric_code`
  - `from` (ISO8601)
  - `to` (ISO8601)
  - Optional tag filters
- **Response 200 (JSON):**
  ```json
  {
    "items": [
      {
        "timestamp": "ISO8601",
        "metric_code": "string",
        "value": 123.45,
        "plot_id": "id",
        "device_id": "id",
        "tags": { "key": "value" }
      }
    ]
  }
  ```

### 3.3 Real-Time Curves and Trends

- **GET** `/api/v1/monitoring/trends`
- **Description:** Return aggregated statistics (daily, weekly, monthly) with year-over-year and month-over-month comparisons.
- **Query Parameters (examples):**
  - `metric_code`
  - `plot_id` or `device_id`
  - `granularity` (e.g. `daily`, `weekly`, `monthly`)
- **Response 200 (JSON):**
  ```json
  {
    "metric_code": "string",
    "granularity": "daily",
    "series": [
      {
        "period_start": "ISO8601",
        "value": 123.45,
        "yoy_change": 0.05,
        "mom_change": -0.01
      }
    ]
  }
  ```

### 3.4 Dashboard Configuration

- **POST** `/api/v1/monitoring/dashboards`
- **Description:** Save a dashboard configuration for a user.
- **Request Body:**
  ```json
  {
    "name": "string",
    "filters": {
      "plots": ["id"],
      "devices": ["id"],
      "metrics": ["metric_code"],
      "time_range": {
        "from": "ISO8601",
        "to": "ISO8601"
      },
      "tags": { "key": "value" }
    }
  }
  ```

- **GET** `/api/v1/monitoring/dashboards`
  - List saved dashboards.

- **GET** `/api/v1/monitoring/dashboards/{id}`
  - Fetch a single dashboard configuration.

### 3.5 Export Monitoring Data

- **GET** `/api/v1/monitoring/export`
- **Description:** Export monitoring data based on filters and time window.
- **Query Parameters:**
  - Same filters as `/monitoring/data`
  - `format` (e.g. `csv` or `json`)
- **Response:**
  - File download in the requested format.

---

## 4. Analysis Domain

### 4.1 Master Data — Plots

- **GET** `/api/v1/analysis/plots`
- **POST** `/api/v1/analysis/plots`
- **GET** `/api/v1/analysis/plots/{id}`
- **PUT** `/api/v1/analysis/plots/{id}`
- **DELETE** `/api/v1/analysis/plots/{id}`

Similar CRUD endpoints exist for:

- `/api/v1/analysis/devices`
- `/api/v1/analysis/metrics`

### 4.2 Indicator Version Management

- **POST** `/api/v1/analysis/metrics/{metric_code}/versions`
- **Description:** Create a new indicator version with meta-information.
- **Request Body (JSON):**
  ```json
  {
    "definition": { "any": "indicator_definition" },
    "diff_description": "string"
  }
  ```
- **Behaviour:**
  - Records modifier (current user) and timestamp automatically.

- **GET** `/api/v1/analysis/metrics/{metric_code}/versions`
  - List versions for the metric, including modifier, timestamp, and difference description.

### 4.3 Analysis Results (Trends, Funnels, Retention)

- **GET** `/api/v1/analysis/trends`
  - Returns indicator-based trend results, with options for drill-down.

- **GET** `/api/v1/analysis/funnels`
  - Returns funnel-style analysis results.

- **GET** `/api/v1/analysis/retention`
  - Returns retention analysis results.

Each of these should support filters by indicator, time range, and relevant dimensions. Exact JSON payload shape must be consistent with the design but must not introduce new business features.

---

## 5. Communication Domain

### 5.1 Create Conversation Message

- **POST** `/api/v1/communication/orders/{order_id}/messages`
- **Description:** Add a message to an order-level conversation.
- **Request Body:**
  ```json
  {
    "content": "string"
  }
  ```
- **Behaviour:**
  - Enforces:
    - Per-user rate limit of 20 messages per minute.
    - Sensitive word interception (blocked messages are not delivered but are logged).

### 5.2 List Conversation Messages

- **GET** `/api/v1/communication/orders/{order_id}/messages`
- **Description:** List messages for the order, including read status.

### 5.3 Update Read Status

- **PUT** `/api/v1/communication/orders/{order_id}/messages/{message_id}/read`
- **Description:** Mark a message as read for the current user.

### 5.4 Transfer Ticket

- **POST** `/api/v1/communication/orders/{order_id}/transfer`
- **Description:** Transfer responsibility for the conversation/ticket.
- **Request Body:**
  ```json
  {
    "target_user_id": "string_or_id"
  }
  ```

---

## 6. Results Management Domain

### 6.1 Create Result

- **POST** `/api/v1/results`
- **Description:** Create a new result entry (paper, project, or patent) in `draft` status.
- **Request Body:**
  ```json
  {
    "type": "paper | project | patent",
    "fields": { "field_name": "value" }
  }
  ```

### 6.2 Update Result Fields

- **PUT** `/api/v1/results/{id}`
- **Description:** Update result fields while in allowed states (e.g. draft, returned).

### 6.3 Change Result Status

- **POST** `/api/v1/results/{id}/status`
- **Description:** Change result status following:
  - `draft → submitted → returned → approved → archived`
- **Request Body:**
  ```json
  {
    "status": "submitted | returned | approved | archived",
    "reason": "optional reason (required for invalidation)"
  }
  ```
- **Behaviour:**
  - Enforce valid transitions only.
  - After archived:
    - Only retrospective notes may be added via a dedicated endpoint.
    - Invalidation requires a reason and is recorded with full audit trace.

### 6.4 Append Retrospective Notes

- **POST** `/api/v1/results/{id}/notes`
- **Description:** Append a note to an archived result.

---

## 7. Evaluation Task Domain

### 7.1 Generate Tasks

- **POST** `/api/v1/tasks/generate`
- **Description:** Generate evaluation tasks based on objects and cycles.
- **Request Body:**
  ```json
  {
    "object_id": "string_or_id",
    "cycle": "string",
    "visibility_start": "ISO8601",
    "visibility_end": "ISO8601"
  }
  ```

### 7.2 List Tasks

- **GET** `/api/v1/tasks`
- **Description:** List tasks filtered by status, object, or cycle.

### 7.3 Submit Task

- **POST** `/api/v1/tasks/{id}/submit`
- **Description:** Submit a task for review.

### 7.4 Review Task

- **POST** `/api/v1/tasks/{id}/review`
- **Description:** Move task to reviewed state, following the prompt’s review node semantics.

### 7.5 Overdue Management

- Overdue logic (default 7 days) is enforced by service logic or background processing and is not a direct endpoint, but task listings reflect delayed status where applicable.

---

## 8. Audit and Capacity Monitoring

### 8.1 Global Audit Log

- **GET** `/api/v1/audit/logs`
- **Description:** Administrative endpoint to view audit logs of operations.
- **Filters:**
  - By user, resource, action, time period.

### 8.2 Capacity Notifications

- **GET** `/api/v1/system/capacity`
- **Description:** View current capacity-related information (e.g., disk usage).
- Threshold logic (e.g., disk > 80%) is implemented internally and triggers notifications or entries accessible through this endpoint.

---

## 9. Health and Utility

### 9.1 Health Check

- **GET** `/api/v1/health`
- **Description:** Returns basic health information indicating that the API and database are reachable.
- **Response:**
  ```json
  {
    "status": "ok"
  }
  ```
