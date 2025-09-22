#!/bin/bash

echo "=== Incident Management System - Setup Verification ==="
echo

# Check project structure
echo "✓ Checking project structure..."

# Backend structure
if [ -d "backend" ]; then
    echo "  ✓ Backend directory exists"
    if [ -f "backend/go.mod" ]; then
        echo "  ✓ Go module file exists"
    else
        echo "  ✗ Go module file missing"
    fi
    if [ -f "backend/main.go" ]; then
        echo "  ✓ Main Go file exists"
    else
        echo "  ✗ Main Go file missing"
    fi
    if [ -d "backend/internal/models" ]; then
        echo "  ✓ Models directory exists"
    else
        echo "  ✗ Models directory missing"
    fi
else
    echo "  ✗ Backend directory missing"
fi

# Frontend structure
if [ -d "frontend" ]; then
    echo "  ✓ Frontend directory exists"
    if [ -f "frontend/package.json" ]; then
        echo "  ✓ Package.json exists"
    else
        echo "  ✗ Package.json missing"
    fi
    if [ -f "frontend/vite.config.ts" ]; then
        echo "  ✓ Vite config exists"
    else
        echo "  ✗ Vite config missing"
    fi
    if [ -d "frontend/src" ]; then
        echo "  ✓ Source directory exists"
    else
        echo "  ✗ Source directory missing"
    fi
else
    echo "  ✗ Frontend directory missing"
fi

# Build scripts
if [ -f "Makefile" ]; then
    echo "  ✓ Makefile exists"
else
    echo "  ✗ Makefile missing"
fi

if [ -f "README.md" ]; then
    echo "  ✓ README.md exists"
else
    echo "  ✗ README.md missing"
fi

echo
echo "=== Setup verification complete ==="
echo
echo "Next steps:"
echo "1. Install Go 1.21+ if not already installed"
echo "2. Install Node.js 18+ if not already installed"
echo "3. Run 'make install' to install dependencies"
echo "4. Run 'make dev' to start development environment"