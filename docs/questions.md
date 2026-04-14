## 1. Authentication Token Format

**Question:**  
The original prompt states that the account system supports only local username + password validation, but does not specify how authenticated sessions should be represented (e.g., session cookies, JWT tokens, or another mechanism).

**Assumption:**  
A single, consistent token-based mechanism is used for authenticated requests (e.g., bearer token in an Authorization header).

**Solution:**  
Implement a login endpoint that returns a token used in an Authorization header for subsequent requests. All protected endpoints validate this token before performing operations. The exact token format is internal to the backend and not exposed as a business feature.

---

## 2. Role-Based Permissions Granularity

**Question:**  
The prompt lists four roles (administrators, researchers, reviewers, customer service) but does not define which endpoints are accessible to which roles in detail.

**Assumption:**  
All authenticated roles can access basic data relevant to their workflows, and administrative operations (such as system-level configuration and audit log viewing) are restricted to administrators.

**Solution:**  
Implement a role attribute on user accounts and enforce authorization checks in the service or middleware layer for operations that should be restricted (e.g., configuration of indicator definitions, access to audit logs, or capacity monitoring endpoints).

---

## 3. Sensitive Word List Management

**Question:**  
The communication domain requires that messages containing sensitive words be intercepted and logged, but the original prompt does not specify how the sensitive word list is defined or maintained.

**Assumption:**  
The sensitive word list is managed as a configurable resource within the system (e.g., stored in the database or configuration) and applies uniformly to all conversation messages.

**Solution:**  
Implement a sensitive word check that reads from a configurable list. During message creation, the content is checked against this list. Messages containing sensitive words are not stored or delivered and are instead logged to an audit or dedicated logging mechanism.

---

## 4. Representation of Cycles in Task Orchestration

**Question:**  
The evaluation task domain specifies that tasks are generated based on “objects and cycles”, but does not define how cycles (such as daily or weekly) are represented or stored.

**Assumption:**  
Cycles are represented as a string or enumerated value indicating the frequency or type of cycle (e.g., “daily”, “weekly”, “monthly”, or a domain-specific label).

**Solution:**  
Introduce a cycle field on the task definition that stores a cycle identifier. The task generation logic uses this identifier to create tasks appropriately, while the overdue logic still uses the default 7-day overdue rule specified by the prompt.

---

## 5. Dashboard Configuration Ownership and Sharing

**Question:**  
The monitoring domain states that users can save custom dashboard conditions, but does not specify whether dashboards are personal or sharable across users.

**Assumption:**  
Dashboards are owned by the user who created them and are not shared by default.

**Solution:**  
Associate each dashboard configuration with the creating user’s account. APIs operate on dashboards scoped to the authenticated user. Any future sharing or multi-user visibility would require a new requirement outside the original prompt.

---

## 6. Result Field Rule Storage

**Question:**  
The results management domain requires configurable field rules (required fields, length limits, enumerations) but does not specify how these rules are stored or versioned.

**Assumption:**  
Field rules are stored in a configuration structure in the database, scoped by result type (paper, project, patent) and used at validation time. Versioning of these rules is not explicitly required beyond what is necessary to apply current rules.

**Solution:**  
Implement a configuration table storing field rules per result type. The results service reads these rules when creating or updating a result and enforces required, length, and enumeration constraints accordingly.

---

## 7. Notification Delivery Mechanism for Capacity Alerts

**Question:**  
The prompt requires capacity monitoring to trigger system notifications when thresholds (e.g., disk usage > 80%) are exceeded, but does not specify how notifications are delivered.

**Assumption:**  
Capacity notifications are internal to the system and do not require integration with external email or SMS providers.

**Solution:**  
Record capacity events in a database table and expose them via a dedicated API endpoint. Administrative users can query and review these notifications. No external messaging integration is implemented.

---

## 8. Overdue Task Delay Marking Detail

**Question:**  
The evaluation task domain states that overdue tasks (default 7 days) are automatically marked as delayed, but does not define whether tasks are considered delayed relative to a due date or another reference point.

**Assumption:**  
Tasks are considered delayed if they remain incomplete 7 days after a defined due datetime or expected completion datetime assigned to the task.

**Solution:**  
Store a due datetime for each task. A background process or scheduled check periodically marks tasks as delayed when the current time exceeds the due datetime by 7 days and the task has not been completed or moved out of the active state.

---

## 9. Export Format Details

**Question:**  
The monitoring domain requires export to local files but does not specify exact formats or file naming conventions.

**Assumption:**  
CSV and JSON are sufficient formats, and file names can be based on timestamp and the type of data exported.

**Solution:**  
Implement endpoints that return data in CSV or JSON based on a format parameter. Use a consistent, time-stamped naming scheme in response headers or metadata to support file downloads, without adding new business semantics beyond the original prompt.

---

## 10. Indicator Analysis Output Shape

**Question:**  
The analysis domain mentions trends, funnels, and retention but does not specify the exact JSON shape for these outputs.

**Assumption:**  
Each analysis endpoint can return structured data tailored to the analysis type, as long as it clearly includes indicator identifiers, time ranges, and values needed for trend, funnel, or retention interpretation.

**Solution:**  
Design response structures that include the minimal fields required to express the requested analysis type (e.g., time buckets and values for trends, step names and counts for funnels, cohort identifiers and retention rates for retention) without adding new business concepts beyond those implied by the original prompt.
