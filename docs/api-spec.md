# Agricultural Research Data & Results Operation Platform — API Specification

This document outlines the main external-facing API endpoints of the platform. It follows the domains described in the original prompt: Monitoring, Analysis, Communication, Results, Evaluation Tasks, Account, and Audit/Capacity.

> NOTE: This is a high-level REST specification based on the original requirements. Field names and exact payloads should be implemented consistently with the design and are not allowed to add new business features.

## 1. Common

- Base URL: `/v1`
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

- **POST** `/v1/auth/register`
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

- **POST** `/v1/auth/login`
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

- **POST** `/v1/monitoring/ingest`
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

- **GET** `/v1/monitoring/data`
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

- **POST** `/v1/monitoring/trends`
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

- **POST** `/v1/dashboards`
- **Description:** Save a dashboard configuration for a user.
- **Request Body:**
  ```json
  {
    "name": "string",
    "config": "JSON string with filter/chart settings"
  }
  ```

- **GET** `/v1/dashboards`
  - List saved dashboards (scoped to authenticated user).

- **GET** `/v1/dashboards/{id}`
  - Fetch a single dashboard configuration (ownership enforced).

- **PUT** `/v1/dashboards/{id}`
  - Update a dashboard configuration.

- **DELETE** `/v1/dashboards/{id}`
  - Delete a dashboard configuration.

### 3.5 Export Monitoring Data

- **GET** `/v1/monitoring/export/json`
  - Export monitoring data as JSON download.
- **GET** `/v1/monitoring/export/csv`
  - Export monitoring data as CSV download.
- **Query Parameters:**
  - Same filters as `/v1/monitoring/data` (plot_id, device_id, metric_code, start_time, end_time, tags).

---

## 4. Analysis Domain

### 4.1 Master Data — Plots

- **POST** `/v1/plots` — Create a plot (admin/researcher)
- **GET** `/v1/plots` — List plots (scoped by user ownership; admin sees all)
- **GET** `/v1/plots/{id}` — Get plot (ownership enforced)
- **PUT** `/v1/plots/{id}` — Update plot (owner/admin)
- **DELETE** `/v1/plots/{id}` — Delete plot (owner/admin)

Similar CRUD endpoints exist for:

- `/v1/devices` — Devices linked to plots (scoped by plot ownership)
- `/v1/metrics` — Basic metric readings

### 4.2 Indicator Version Management

- **POST** `/v1/indicators` — Create a new indicator definition (admin/researcher).
- **GET** `/v1/indicators` — List indicator definitions.
- **GET** `/v1/indicators/{id}` — Get indicator definition.
- **PUT** `/v1/indicators/{id}` — Update indicator and record a new version with diff (admin/researcher).
  - **Request Body:**
    ```json
    {
      "name": "string",
      "description": "string",
      "unit": "string",
      "formula": "string",
      "category": "string",
      "diff_summary": "Description of what changed (required)"
    }
    ```
  - Records modifier (current user) and timestamp automatically.
- **DELETE** `/v1/indicators/{id}` — Deprecate an indicator (admin only).
- **GET** `/v1/indicators/{id}/versions` — List all versions with modifier, timestamp, and diff.
- **GET** `/v1/indicators/{id}/versions/{version}` — Get a specific version.

### 4.3 Analysis Results (Trends, Funnels, Retention)

- **POST** `/v1/analysis/trends`
  - Returns indicator-based trend results, with options for drill-down by plot_id or device_id.

- **POST** `/v1/analysis/funnels`
  - Returns funnel-style analysis results across sequential metric stages.

- **POST** `/v1/analysis/retention`
  - Returns cohort-based retention analysis results.

Each supports filters by metric_code, time range, plot_id, device_id, and drill_by dimensions.

---

## 5. Communication Domain

### 5.1 Orders

- **POST** `/v1/orders` — Create a new order (admin/researcher/reviewer/customer_service).
- **GET** `/v1/orders` — List orders (scoped to user's own orders or assignments).
- **GET** `/v1/orders/{id}` — Get order details (ownership enforced).

### 5.2 Create Conversation Message

- **POST** `/v1/orders/{order_id}/messages`
- **Description:** Add a message to an order-level conversation.
- **Request Body:**
  ```json
  {
    "message": "string"
  }
  ```
- **Behaviour:**
  - Enforces:
    - Per-user rate limit of 20 messages per minute.
    - Sensitive word interception (blocked messages are not delivered but are logged).

### 5.3 List Conversation Messages

- **GET** `/v1/orders/{order_id}/messages`
- **Description:** List messages for the order (ownership enforced), including read status.

### 5.4 Update Read Status

- **PATCH** `/v1/orders/{order_id}/messages/{message_id}/read`
- **Description:** Mark a message as read for the current user.

### 5.5 Transfer Ticket

- **POST** `/v1/orders/{order_id}/transfer`
- **Description:** Transfer responsibility for the conversation/ticket (ownership enforced).
- **Request Body:**
  ```json
  {
    "transfer_to_user_id": 123,
    "reason": "optional reason"
  }
  ```

### 5.6 Templates

- **POST** `/v1/templates` — Create message template (admin/customer_service).
- **GET** `/v1/templates` — List templates.
- **POST** `/v1/orders/{order_id}/templates/{template_id}` — Send template into order conversation.

---

## 6. Results Management Domain

### 6.1 Create Result

- **POST** `/v1/results`
- **Description:** Create a new result entry (paper, project, or patent) in `draft` status.
- **Request Body:**
  ```json
  {
    "type": "paper | project | patent",
    "fields": { "field_name": "value" }
  }
  ```

### 6.2 Update Result Fields

- **PUT** `/v1/results/{id}`
- **Description:** Update result fields while in allowed states (e.g. draft, returned).

### 6.3 Change Result Status

- **PATCH** `/v1/results/{id}/transition`
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

- **POST** `/v1/results/{id}/notes`
- **Description:** Append a note to an archived result.

---

## 7. Evaluation Task Domain

### 7.1 Generate Tasks

- **POST** `/v1/tasks/generate`
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

- **GET** `/v1/tasks`
- **Description:** List tasks filtered by status, object, or cycle.

### 7.3 Submit Task

- **POST** `/v1/tasks/{id}/submit`
- **Description:** Submit a task for review.

### 7.4 Review Task

- **POST** `/v1/tasks/{id}/review`
- **Description:** Move task to reviewed state, following the prompt’s review node semantics.

### 7.5 Overdue Management

- Overdue logic (default 7 days) is enforced by service logic or background processing and is not a direct endpoint, but task listings reflect delayed status where applicable.

---

## 8. Audit and Capacity Monitoring

### 8.1 Global Audit Log

- **GET** `/v1/audit/logs`
- **Description:** Administrative endpoint to view audit logs of operations.
- **Filters:**
  - By user, resource, action, time period.

### 8.2 Capacity Notifications

- **GET** `/v1/system/capacity`
- **Description:** View current capacity-related information (e.g., disk usage).
- Threshold logic (e.g., disk > 80%) is implemented internally and triggers notifications or entries accessible through this endpoint.

---

## 9. Health and Utility

### 9.1 Health Check

- **GET** `/v1/health`
- **Description:** Returns basic health information indicating that the API and database are reachable.
- **Response:**
  ```json
  {
    "status": "ok"
  }
  ```
