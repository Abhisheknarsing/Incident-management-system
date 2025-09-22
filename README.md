# Incident Management System

A comprehensive incident management and analytics platform built with Go backend and React frontend.

## Features

- **File Upload**: Upload Excel files containing incident data
- **Data Processing**: Automated processing with sentiment analysis and automation opportunity identification
- **Analytics Dashboard**: Interactive visualizations for incident trends, priorities, applications, and resolution metrics
- **Export Functionality**: Export data and analytics in multiple formats
- **Performance Monitoring**: Built-in performance and memory monitoring
- **Error Handling**: Comprehensive error handling and logging

## Technology Stack

### Backend
- **Language**: Go
- **Framework**: Gin
- **Database**: SQLite (with potential for PostgreSQL/MySQL)
- **Excel Processing**: Custom Excel parser
- **Caching**: Ristretto
- **Logging**: Structured JSON logging

### Frontend
- **Framework**: React with TypeScript
- **Build Tool**: Vite
- **UI Library**: Tailwind CSS with custom components
- **Charts**: Recharts
- **State Management**: Zustand
- **HTTP Client**: Axios
- **Data Fetching**: React Query

## Prerequisites

- Go 1.19+
- Node.js 16+
- npm 8+
- Git

## Quick Start

### Backend Setup
```bash
# Clone the repository
git clone https://github.com/your-username/incident-management-system.git
cd incident-management-system/backend

# Install dependencies
go mod tidy

# Build and run
go build -o main .
./main
```

The backend server will start on port 8080.

### Frontend Setup
```bash
# Navigate to frontend directory
cd ../frontend

# Install dependencies
npm install

# Start development server
npm run dev
```

The frontend development server will start on port 5173 (or another available port).

## Project Structure

```
incident-management-system/
├── backend/
│   ├── cmd/
│   ├── internal/
│   │   ├── handlers/
│   │   ├── services/
│   │   ├── database/
│   │   ├── models/
│   │   ├── logging/
│   │   ├── monitoring/
│   │   ├── storage/
│   │   └── errors/
│   ├── uploads/
│   └── main.go
├── frontend/
│   ├── src/
│   │   ├── components/
│   │   ├── hooks/
│   │   ├── lib/
│   │   ├── pages/
│   │   ├── types/
│   │   └── App.tsx
│   ├── public/
│   └── package.json
├── sample_data/
│   ├── incidents_sample.csv
│   ├── incidents_sample.xlsx
│   ├── incidents_large_sample.csv
│   └── incidents_large_sample.xlsx
├── docs/
│   ├── user_guide.md
│   ├── api_documentation.md
│   └── deployment_guide.md
├── scripts/
│   └── csv_to_excel.py
└── README.md
```

## API Documentation

Detailed API documentation is available in [docs/api_documentation.md](docs/api_documentation.md).

## Sample Data

The project includes sample data files for testing:
- [sample_data/incidents_sample.xlsx](sample_data/incidents_sample.xlsx) - Small sample dataset
- [sample_data/incidents_large_sample.xlsx](sample_data/incidents_large_sample.xlsx) - Larger sample dataset

## Documentation

- [User Guide](docs/user_guide.md) - Instructions for using the application
- [API Documentation](docs/api_documentation.md) - Detailed API endpoints and usage
- [Deployment Guide](docs/deployment_guide.md) - Instructions for deploying to production

## Key Components

### Backend Services
1. **Upload Service**: Handles file uploads and storage
2. **Processing Service**: Processes Excel files and performs analysis
3. **Analytics Service**: Generates incident analytics and metrics
4. **Export Service**: Handles data export functionality
5. **Monitoring Service**: Performance and memory monitoring
6. **Logging Service**: Structured logging with different levels

### Frontend Components
1. **Upload Page**: File upload interface with drag-and-drop
2. **Processing Page**: Data processing management and status tracking
3. **Dashboard Page**: Analytics visualizations and metrics
4. **Custom Charts**: Interactive charts for data visualization
5. **Filter System**: Advanced filtering capabilities
6. **Export Functionality**: Data export in multiple formats

## Performance Optimizations

- Database query caching with Ristretto
- API response caching
- Concurrent processing optimizations
- Memory usage monitoring
- Code splitting and lazy loading
- Virtual scrolling for large data tables
- Bundle size optimization

## Error Handling

- Structured logging with different log levels
- Error middleware for API endpoints
- Standardized error response formats
- Error tracking and monitoring
- User-friendly error message display
- Network error handling with retry mechanisms

## Testing

The application includes comprehensive testing:
- Unit tests for backend services
- Integration tests for API endpoints
- Component tests for frontend interfaces
- End-to-end user workflow testing

## Contributing

1. Fork the repository
2. Create a feature branch
3. Commit your changes
4. Push to the branch
5. Create a pull request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

For support, please open an issue on the GitHub repository or contact the maintainers.