# Incident Management System

A full-stack web application for processing incident data from Excel uploads and providing comprehensive analytics through an interactive dashboard.

## Technology Stack

### Backend
- **Go 1.21+** with Gin framework
- **DuckDB** for analytical data storage
- **Excelize** library for Excel file processing
- RESTful API design

### Frontend
- **React 18+** with Vite
- **ShadCN UI** component library
- **Recharts** for data visualization
- **React Query** for API state management
- **Tailwind CSS** for styling

## Project Structure

```
├── backend/                 # Go backend application
│   ├── internal/
│   │   ├── models/         # Data models and structures
│   │   ├── handlers/       # HTTP request handlers
│   │   ├── services/       # Business logic services
│   │   └── database/       # Database connection and queries
│   ├── uploads/            # Uploaded Excel files storage
│   ├── go.mod              # Go module dependencies
│   └── main.go             # Application entry point
├── frontend/               # React frontend application
│   ├── src/
│   │   ├── components/     # Reusable UI components
│   │   ├── pages/          # Page components
│   │   ├── lib/            # Utility functions and API client
│   │   └── types/          # TypeScript type definitions
│   ├── package.json        # Node.js dependencies
│   └── vite.config.ts      # Vite configuration
├── Makefile               # Build and development scripts
└── README.md              # Project documentation
```

## Getting Started

### Prerequisites
- Go 1.21 or higher
- Node.js 18 or higher
- npm or yarn

### Installation

1. Clone the repository
2. Install dependencies:
   ```bash
   make install
   ```

### Development

Start both backend and frontend in development mode:
```bash
make dev
```

Or start them separately:
```bash
# Backend only (http://localhost:8080)
make backend-dev

# Frontend only (http://localhost:5173)
make frontend-dev
```

### Building for Production

```bash
make build
```

### Available Commands

- `make install` - Install all dependencies
- `make dev` - Start development environment
- `make build` - Build for production
- `make test` - Run tests
- `make clean` - Clean build artifacts

## Features

### Data Upload
- Excel file upload (.xlsx, .xls)
- File validation and error reporting
- Upload progress tracking

### Data Processing
- Asynchronous data processing
- Sentiment analysis of incident descriptions
- Automation opportunity identification
- Progress monitoring

### Analytics Dashboard
- Timeline visualization of incidents
- Priority analysis and distribution
- Application-wise incident breakdown
- Sentiment analysis results
- Resolution metrics and trends
- Automation opportunities analysis

### Export and Filtering
- Multi-format export (CSV, PDF)
- Real-time filtering capabilities
- Date range, priority, and application filters

## API Endpoints

### Upload Management
- `POST /api/uploads` - Upload Excel file
- `GET /api/uploads` - List uploads with status
- `POST /api/uploads/:id/analyze` - Start analysis
- `GET /api/uploads/:id/status` - Get processing status

### Analytics
- `GET /api/analytics/timeline` - Timeline data
- `GET /api/analytics/priorities` - Priority analysis
- `GET /api/analytics/applications` - Application analysis
- `GET /api/analytics/sentiment` - Sentiment analysis
- `GET /api/analytics/resolution` - Resolution metrics
- `GET /api/analytics/automation` - Automation opportunities

### Export
- `POST /api/export` - Export data in various formats

## Development Status

This project is currently in the initial setup phase. The basic project structure and dependencies have been configured. Upcoming development tasks include:

1. Database setup and schema creation
2. Excel file processing implementation
3. Data analysis and processing engine
4. Analytics API development
5. Frontend dashboard implementation
6. Export functionality
7. Testing and optimization

## Contributing

Please refer to the task list in `.kiro/specs/incident-management-system/tasks.md` for current development priorities and implementation details.