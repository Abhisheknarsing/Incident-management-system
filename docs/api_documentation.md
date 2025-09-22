# Incident Management System - API Documentation

## Base URL
```
http://localhost:8080/api
```

## Authentication
No authentication required for current version.

## Error Responses
All error responses follow this format:
```json
{
  "code": "ERROR_CODE",
  "message": "Technical error message",
  "user_message": "User-friendly error message",
  "details": "Additional details about the error",
  "timestamp": "2025-09-22T10:00:00Z",
  "request_id": "unique-request-id",
  "path": "/api/endpoint",
  "method": "GET"
}
```

## Upload Endpoints

### Upload File
**POST** `/uploads`

Upload an Excel file containing incident data.

#### Request
- Content-Type: `multipart/form-data`
- Form field: `file` (Excel file)

#### Response
```json
{
  "message": "File uploaded successfully",
  "upload": {
    "id": "uuid",
    "filename": "stored_filename.xlsx",
    "original_filename": "user_provided_filename.xlsx",
    "status": "uploaded",
    "record_count": 0,
    "processed_count": 0,
    "error_count": 0,
    "errors": [],
    "created_at": "2025-09-22T10:00:00Z"
  }
}
```

#### Errors
- `MISSING_FILE`: No file provided
- `FILE_TOO_LARGE`: File exceeds 50MB limit
- `INVALID_FORMAT`: File is not a valid Excel format

### Get All Uploads
**GET** `/uploads`

Retrieve a list of all uploaded files.

#### Response
```json
{
  "uploads": [
    {
      "id": "uuid",
      "filename": "stored_filename.xlsx",
      "original_filename": "user_provided_filename.xlsx",
      "status": "uploaded|processing|completed|failed",
      "record_count": 100,
      "processed_count": 95,
      "error_count": 5,
      "errors": [],
      "created_at": "2025-09-22T10:00:00Z",
      "processed_at": "2025-09-22T10:05:00Z"
    }
  ]
}
```

### Get Specific Upload
**GET** `/uploads/{id}`

Retrieve details for a specific upload.

#### Response
```json
{
  "upload": {
    "id": "uuid",
    "filename": "stored_filename.xlsx",
    "original_filename": "user_provided_filename.xlsx",
    "status": "uploaded|processing|completed|failed",
    "record_count": 100,
    "processed_count": 95,
    "error_count": 5,
    "errors": [],
    "created_at": "2025-09-22T10:00:00Z",
    "processed_at": "2025-09-22T10:05:00Z"
  }
}
```

#### Errors
- `NOT_FOUND`: Upload with specified ID not found

### Start Analysis
**POST** `/uploads/{id}/analyze`

Start processing an uploaded file.

#### Response
```json
{
  "message": "Processing started",
  "upload_id": "uuid"
}
```

#### Errors
- `NOT_FOUND`: Upload with specified ID not found
- `INVALID_STATUS`: Upload is not in a valid state for processing

### Get Processing Status
**GET** `/uploads/{id}/status`

Get the current processing status of an upload.

#### Response
```json
{
  "status": {
    "upload_id": "uuid",
    "status": "pending|processing|completed|failed",
    "total_rows": 100,
    "processed_rows": 75,
    "valid_rows": 95,
    "error_count": 5,
    "errors": ["Error message 1", "Error message 2"],
    "start_time": "2025-09-22T10:00:00Z",
    "end_time": "2025-09-22T10:05:00Z",
    "duration": "5m0s"
  }
}
```

## Analytics Endpoints

### Get Daily Timeline
**GET** `/analytics/timeline/daily`

Get incident timeline data grouped by day.

#### Query Parameters
- `start_date`: Start date (YYYY-MM-DD)
- `end_date`: End date (YYYY-MM-DD)
- `priorities`: Comma-separated list of priorities (P1,P2,P3,P4)
- `applications`: Comma-separated list of applications
- `statuses`: Comma-separated list of statuses

#### Response
```json
{
  "data": [
    {
      "date": "2025-09-22",
      "count": 15,
      "p1_count": 3,
      "p2_count": 5,
      "p3_count": 4,
      "p4_count": 3
    }
  ],
  "filters": {},
  "count": 1
}
```

### Get Weekly Timeline
**GET** `/analytics/timeline/weekly`

Get incident timeline data grouped by week.

#### Query Parameters
- `start_date`: Start date (YYYY-MM-DD)
- `end_date`: End date (YYYY-MM-DD)
- `priorities`: Comma-separated list of priorities (P1,P2,P3,P4)
- `applications`: Comma-separated list of applications
- `statuses`: Comma-separated list of statuses

#### Response
```json
{
  "data": [
    {
      "date": "2025-09-15",
      "count": 45,
      "p1_count": 8,
      "p2_count": 15,
      "p3_count": 12,
      "p4_count": 10
    }
  ],
  "filters": {},
  "count": 1
}
```

### Get Trend Analysis
**GET** `/analytics/trends`

Get incident trend analysis.

#### Query Parameters
- `period`: daily|weekly
- `start_date`: Start date (YYYY-MM-DD)
- `end_date`: End date (YYYY-MM-DD)
- `priorities`: Comma-separated list of priorities (P1,P2,P3,P4)
- `applications`: Comma-separated list of applications
- `statuses`: Comma-separated list of statuses

#### Response
```json
{
  "data": [
    {
      "date": "2025-09-22",
      "count": 15,
      "trend": "stable|up|down",
      "percentage_change": 0.0
    }
  ],
  "period": "daily",
  "filters": {},
  "count": 1
}
```

### Get Priority Analysis
**GET** `/analytics/priority`

Get incident distribution by priority.

#### Query Parameters
- `start_date`: Start date (YYYY-MM-DD)
- `end_date`: End date (YYYY-MM-DD)
- `priorities`: Comma-separated list of priorities (P1,P2,P3,P4)
- `applications`: Comma-separated list of applications
- `statuses`: Comma-separated list of statuses

#### Response
```json
{
  "data": [
    {
      "priority": "P1",
      "count": 25,
      "percentage": 25.0
    }
  ],
  "filters": {},
  "count": 4
}
```

### Get Application Analysis
**GET** `/analytics/applications`

Get incident analysis by application.

#### Query Parameters
- `start_date`: Start date (YYYY-MM-DD)
- `end_date`: End date (YYYY-MM-DD)
- `priorities`: Comma-separated list of priorities (P1,P2,P3,P4)
- `applications`: Comma-separated list of applications
- `statuses`: Comma-separated list of statuses

#### Response
```json
{
  "data": [
    {
      "application_name": "Database Service",
      "incident_count": 30,
      "avg_resolution_time": 120.5,
      "trend": "stable|up|down"
    }
  ],
  "filters": {},
  "count": 5
}
```

### Get Sentiment Analysis
**GET** `/analytics/sentiment`

Get sentiment analysis of incident descriptions.

#### Query Parameters
- `start_date`: Start date (YYYY-MM-DD)
- `end_date`: End date (YYYY-MM-DD)
- `priorities`: Comma-separated list of priorities (P1,P2,P3,P4)
- `applications`: Comma-separated list of applications
- `statuses`: Comma-separated list of statuses

#### Response
```json
{
  "data": {
    "positive": 15,
    "negative": 25,
    "neutral": 60
  },
  "filters": {}
}
```

### Get Resolution Analysis
**GET** `/analytics/resolution`

Get resolution time metrics.

#### Query Parameters
- `start_date`: Start date (YYYY-MM-DD)
- `end_date`: End date (YYYY-MM-DD)
- `priorities`: Comma-separated list of priorities (P1,P2,P3,P4)
- `applications`: Comma-separated list of applications
- `statuses`: Comma-separated list of statuses

#### Response
```json
{
  "data": {
    "avg_resolution_time": 180.5,
    "median_resolution_time": 120.0,
    "resolution_trends": [
      {
        "date": "2025-09-22",
        "count": 15,
        "p1_count": 3,
        "p2_count": 5,
        "p3_count": 4,
        "p4_count": 3
      }
    ]
  },
  "filters": {}
}
```

### Get Automation Analysis
**GET** `/analytics/automation`

Get automation opportunity analysis.

#### Query Parameters
- `start_date`: Start date (YYYY-MM-DD)
- `end_date`: End date (YYYY-MM-DD)
- `priorities`: Comma-separated list of priorities (P1,P2,P3,P4)
- `applications`: Comma-separated list of applications
- `statuses`: Comma-separated list of statuses

#### Response
```json
{
  "data": [
    {
      "it_process_group": "Database Maintenance",
      "automation_score": 0.85,
      "incident_count": 25,
      "automation_feasible": true
    }
  ],
  "filters": {},
  "count": 8
}
```

### Get Dashboard Summary
**GET** `/analytics/summary`

Get a summary of all analytics metrics.

#### Query Parameters
- `start_date`: Start date (YYYY-MM-DD)
- `end_date`: End date (YYYY-MM-DD)
- `priorities`: Comma-separated list of priorities (P1,P2,P3,P4)
- `applications`: Comma-separated list of applications
- `statuses`: Comma-separated list of statuses

#### Response
```json
{
  "data": {
    "timeline": [...],
    "priorities": [...],
    "applications": [...],
    "sentiment": {...},
    "resolutionMetrics": {...},
    "automationOpportunities": [...]
  },
  "filters": {}
}
```

## Export Endpoints

### Request Export
**POST** `/export`

Request data export in specified format.

#### Request Body
```json
{
  "format": "csv|pdf",
  "dataType": "all|timeline|priority|application|sentiment|resolution|automation",
  "filters": {
    "start_date": "2025-09-01",
    "end_date": "2025-09-30",
    "priorities": ["P1", "P2"],
    "applications": ["Database Service", "API Gateway"]
  }
}
```

#### Response
```json
{
  "download_url": "/api/export/download/job-id",
  "job_id": "export-job-id"
}
```

### Get Export Status
**GET** `/export/{job_id}/status`

Get the status of an export job.

#### Response
```json
{
  "id": "export-job-id",
  "status": "pending|processing|completed|failed",
  "progress": 75,
  "download_url": "/api/export/download/job-id",
  "error": "Error message if failed"
}
```

### Download Export
**GET** `/export/download/{job_id}`

Download the exported data file.

#### Response
Binary file content with appropriate Content-Type header.