# Incident Management System - Build and Development Scripts

.PHONY: help install dev build clean test backend-dev frontend-dev

# Default target
help:
	@echo "Available commands:"
	@echo "  install      - Install all dependencies (backend and frontend)"
	@echo "  dev          - Start both backend and frontend in development mode"
	@echo "  backend-dev  - Start only the backend server"
	@echo "  frontend-dev - Start only the frontend development server"
	@echo "  build        - Build both backend and frontend for production"
	@echo "  test         - Run tests for both backend and frontend"
	@echo "  clean        - Clean build artifacts"

# Install dependencies
install:
	@echo "Installing backend dependencies..."
	cd backend && go mod tidy
	@echo "Installing frontend dependencies..."
	cd frontend && npm install

# Development mode - start both services
dev:
	@echo "Starting development environment..."
	@echo "Backend will run on http://localhost:8080"
	@echo "Frontend will run on http://localhost:5173"
	@make -j2 backend-dev frontend-dev

# Start backend development server
backend-dev:
	@echo "Starting backend development server..."
	cd backend && go run main.go

# Start frontend development server
frontend-dev:
	@echo "Starting frontend development server..."
	cd frontend && npm run dev

# Build for production
build:
	@echo "Building backend..."
	cd backend && go build -o bin/incident-management-system main.go
	@echo "Building frontend..."
	cd frontend && npm run build

# Run tests
test:
	@echo "Running backend tests..."
	cd backend && go test ./...
	@echo "Frontend linting skipped - will be configured in later tasks"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf backend/bin
	rm -rf frontend/dist
	rm -rf backend/uploads/*
	@echo "Clean complete"