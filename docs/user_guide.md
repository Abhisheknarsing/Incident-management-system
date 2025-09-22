# Incident Management System - User Guide

## Table of Contents
1. [Introduction](#introduction)
2. [System Requirements](#system-requirements)
3. [Installation](#installation)
4. [Getting Started](#getting-started)
5. [Uploading Incident Data](#uploading-incident-data)
6. [Processing Data](#processing-data)
7. [Viewing Analytics](#viewing-analytics)
8. [Exporting Data](#exporting-data)
9. [Troubleshooting](#troubleshooting)

## Introduction

The Incident Management System is a comprehensive tool for analyzing IT incident data. It allows users to upload Excel files containing incident information, process the data with advanced analytics, and visualize key metrics through interactive dashboards.

## System Requirements

### Backend Requirements
- Go 1.19 or higher
- SQLite database
- 4GB RAM minimum
- 10GB free disk space

### Frontend Requirements
- Node.js 16 or higher
- npm 8 or higher
- Modern web browser (Chrome, Firefox, Safari, Edge)

## Installation

### Backend Installation
1. Clone the repository:
   ```bash
   git clone https://github.com/your-username/incident-management-system.git
   cd incident-management-system/backend
   ```

2. Install Go dependencies:
   ```bash
   go mod tidy
   ```

3. Build the application:
   ```bash
   go build -o main .
   ```

### Frontend Installation
1. Navigate to the frontend directory:
   ```bash
   cd incident-management-system/frontend
   ```

2. Install dependencies:
   ```bash
   npm install
   ```

## Getting Started

### Starting the Backend Server
1. Navigate to the backend directory:
   ```bash
   cd incident-management-system/backend
   ```

2. Run the server:
   ```bash
   ./main
   ```
   The server will start on port 8080.

### Starting the Frontend Application
1. Navigate to the frontend directory:
   ```bash
   cd incident-management-system/frontend
   ```

2. Start the development server:
   ```bash
   npm run dev
   ```
   The application will be available at http://localhost:5173 (or another port if 5173 is in use).

## Uploading Incident Data

1. Navigate to the Upload page in the web application.
2. Click the "Select File" button or drag and drop an Excel file containing incident data.
3. Supported file formats: .xlsx, .xls
4. Maximum file size: 50MB
5. Click "Upload File" to upload the data.

### Required Columns
The Excel file must contain the following columns:
- `incident_id`: Unique identifier for the incident
- `report_date`: Date when the incident was reported (YYYY-MM-DD)
- `brief_description`: Short description of the incident
- `application_name`: Name of the affected application
- `resolution_group`: Team responsible for resolution
- `resolved_person`: Person who resolved the incident
- `priority`: Incident priority (P1, P2, P3, P4)

### Optional Columns
- `resolve_date`: Date when the incident was resolved
- `last_resolve_date`: Last date of resolution (for recurring incidents)
- `description`: Detailed description of the incident
- `category`: Incident category
- `subcategory`: Incident subcategory
- `impact`: Impact level (High, Medium, Low)
- `urgency`: Urgency level (High, Medium, Low)
- `status`: Current status of the incident

## Processing Data

1. After uploading a file, navigate to the Processing page.
2. Find your uploaded file in the list.
3. Click "Analyze Data" to start processing.
4. The system will process the data and generate analytics including:
   - Sentiment analysis of incident descriptions
   - Automation opportunity identification
   - Resolution time calculations
   - Priority distribution analysis

## Viewing Analytics

1. Navigate to the Dashboard page after processing data.
2. View various analytics visualizations:
   - Incident timeline
   - Priority distribution
   - Application analysis
   - Resolution metrics
   - Sentiment analysis
   - Automation opportunities

### Filtering Data
Use the filter panel to narrow down the data by:
- Date range
- Priority levels
- Applications
- Status

## Exporting Data

1. On the Dashboard page, use the "Export Dashboard" button to export all analytics.
2. Individual charts can be exported using the export button in each chart container.
3. Supported export formats:
   - CSV
   - PDF

## Troubleshooting

### Common Issues

**Issue: Upload fails with "File too large"**
Solution: Ensure your file is under 50MB. Consider splitting large files into smaller chunks.

**Issue: Processing takes too long**
Solution: Large files with many records may take several minutes to process. Be patient and avoid refreshing the page.

**Issue: Charts not displaying data**
Solution: Ensure data has been processed successfully. Check the Processing page for any errors.

**Issue: Backend server fails to start**
Solution: Check that port 8080 is not in use by another application.

### Getting Help
For additional support, please contact the system administrator or check the project documentation.