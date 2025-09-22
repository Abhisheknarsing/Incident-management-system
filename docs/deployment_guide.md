# Incident Management System - Deployment Guide

## Table of Contents
1. [Prerequisites](#prerequisites)
2. [Backend Deployment](#backend-deployment)
3. [Frontend Deployment](#frontend-deployment)
4. [Database Configuration](#database-configuration)
5. [Environment Variables](#environment-variables)
6. [Production Considerations](#production-considerations)
7. [Monitoring and Maintenance](#monitoring-and-maintenance)

## Prerequisites

### System Requirements
- Linux/Unix server or Windows Server
- 4+ CPU cores
- 8GB+ RAM
- 20GB+ free disk space
- Internet access for package installation

### Software Dependencies
- Go 1.19+
- Node.js 16+
- npm 8+
- SQLite 3.35+
- Git

## Backend Deployment

### 1. Clone the Repository
```bash
git clone https://github.com/your-username/incident-management-system.git
cd incident-management-system/backend
```

### 2. Build the Application
```bash
# Download dependencies
go mod tidy

# Build the binary
go build -o incident-management-system .
```

### 3. Create Deployment Directory
```bash
# Create deployment directory
sudo mkdir -p /opt/incident-management-system

# Copy binary and assets
sudo cp incident-management-system /opt/incident-management-system/
sudo cp -r uploads /opt/incident-management-system/
sudo mkdir -p /opt/incident-management-system/logs
```

### 4. Create Systemd Service (Linux)
Create a systemd service file at `/etc/systemd/system/incident-management-system.service`:

```ini
[Unit]
Description=Incident Management System Backend
After=network.target

[Service]
Type=simple
User=ims-user
WorkingDirectory=/opt/incident-management-system
ExecStart=/opt/incident-management-system/incident-management-system
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
```

### 5. Create Service User
```bash
sudo useradd -r -s /bin/false ims-user
sudo chown -R ims-user:ims-user /opt/incident-management-system
```

### 6. Start the Service
```bash
sudo systemctl daemon-reload
sudo systemctl enable incident-management-system
sudo systemctl start incident-management-system
sudo systemctl status incident-management-system
```

## Frontend Deployment

### 1. Build the Frontend
```bash
# Navigate to frontend directory
cd ../frontend

# Install dependencies
npm install

# Build for production
npm run build
```

### 2. Deploy to Web Server
The build output will be in the `dist` directory. Copy this to your web server's document root.

#### Using Nginx
Create an Nginx configuration file at `/etc/nginx/sites-available/incident-management-system`:

```nginx
server {
    listen 80;
    server_name your-domain.com;
    
    root /var/www/incident-management-system/dist;
    index index.html;
    
    # Serve static files
    location / {
        try_files $uri $uri/ /index.html;
    }
    
    # Proxy API requests to backend
    location /api/ {
        proxy_pass http://localhost:8080/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
    
    # Security headers
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header Referrer-Policy "no-referrer-when-downgrade" always;
    add_header Content-Security-Policy "default-src 'self' http: https: data: blob: 'unsafe-inline'" always;
}
```

Enable the site:
```bash
sudo ln -s /etc/nginx/sites-available/incident-management-system /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

#### Using Apache
Create an Apache virtual host configuration:

```apache
<VirtualHost *:80>
    ServerName your-domain.com
    DocumentRoot /var/www/incident-management-system/dist
    
    # Serve static files
    <Directory /var/www/incident-management-system/dist>
        AllowOverride All
        Require all granted
        RewriteEngine On
        RewriteBase /
        RewriteRule ^index\.html$ - [L]
        RewriteCond %{REQUEST_FILENAME} !-f
        RewriteCond %{REQUEST_FILENAME} !-d
        RewriteRule . /index.html [L]
    </Directory>
    
    # Proxy API requests to backend
    ProxyPreserveHost On
    ProxyPass /api/ http://localhost:8080/
    ProxyPassReverse /api/ http://localhost:8080/
</VirtualHost>
```

## Database Configuration

### SQLite Database
The application uses SQLite by default. The database file is created automatically at `incident_management.db`.

### Database Backup Strategy
```bash
# Create backup script
cat > /opt/incident-management-system/backup-db.sh << 'EOF'
#!/bin/bash
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_DIR="/opt/incident-management-system/backups"
DB_FILE="/opt/incident-management-system/incident_management.db"

mkdir -p $BACKUP_DIR
cp $DB_FILE $BACKUP_DIR/incident_management_$DATE.db
gzip $BACKUP_DIR/incident_management_$DATE.db

# Keep only last 30 days of backups
find $BACKUP_DIR -name "incident_management_*.db.gz" -mtime +30 -delete
EOF

chmod +x /opt/incident-management-system/backup-db.sh

# Add to crontab for daily backups at 2 AM
echo "0 2 * * * /opt/incident-management-system/backup-db.sh" | crontab -
```

## Environment Variables

### Backend Environment Variables
Create a `.env` file in the backend directory:

```bash
# Database configuration
DB_PATH=/opt/incident-management-system/incident_management.db

# Logging configuration
LOG_LEVEL=info
LOG_FORMAT=json
LOG_OUTPUT=file
LOG_FILE=/opt/incident-management-system/logs/backend.log

# Performance monitoring
MONITORING_ENABLED=true
```

### Frontend Environment Variables
Create a `.env.production` file in the frontend directory:

```bash
# API configuration
VITE_API_URL=https://your-domain.com/api

# Frontend settings
VITE_APP_NAME=Incident Management System
VITE_APP_VERSION=1.0.0
```

## Production Considerations

### Security
1. Use HTTPS in production
2. Implement proper authentication and authorization
3. Regularly update dependencies
4. Restrict file upload types and sizes
5. Implement rate limiting
6. Use a firewall to restrict access

### Performance
1. Use a reverse proxy (Nginx/Apache) for static file serving
2. Enable gzip compression
3. Configure caching headers
4. Monitor memory usage
5. Implement database query optimization

### Scalability
1. Consider using a more robust database (PostgreSQL, MySQL) for high-volume deployments
2. Implement load balancing for multiple backend instances
3. Use a CDN for static assets
4. Implement horizontal scaling for processing jobs

### High Availability
1. Use a process manager like PM2 for the backend
2. Implement database replication
3. Use redundant servers
4. Implement automatic failover

## Monitoring and Maintenance

### Health Checks
The application provides health check endpoints:
- `/health`: Overall system health
- `/metrics`: Performance metrics
- `/memory`: Memory usage information

### Log Management
```bash
# View backend logs
journalctl -u incident-management-system -f

# View frontend logs (if using PM2)
pm2 logs incident-management-frontend
```

### Monitoring Scripts
Create a monitoring script to check system health:

```bash
cat > /opt/incident-management-system/monitor.sh << 'EOF'
#!/bin/bash
HEALTH_CHECK_URL="http://localhost:8080/health"
RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" $HEALTH_CHECK_URL)

if [ $RESPONSE -eq 200 ]; then
    echo "System is healthy"
    exit 0
else
    echo "System health check failed with status $RESPONSE"
    # Send alert (email, Slack, etc.)
    exit 1
fi
EOF

chmod +x /opt/incident-management-system/monitor.sh

# Add to crontab for regular monitoring
echo "*/5 * * * * /opt/incident-management-system/monitor.sh" | crontab -
```

### Maintenance Tasks
1. Regular database backups
2. Log rotation
3. Dependency updates
4. Security patches
5. Performance monitoring
6. User access reviews

### Troubleshooting
Common issues and solutions:

1. **Service won't start**: Check logs with `journalctl -u incident-management-system`
2. **Database corruption**: Restore from latest backup
3. **High memory usage**: Check for memory leaks, restart service if needed
4. **Slow performance**: Check database queries, optimize as needed
5. **File upload issues**: Check disk space, permissions

### Backup and Recovery
1. Database backups (automated daily)
2. File backups (uploaded files)
3. Configuration backups
4. Recovery procedures documented

This deployment guide provides a comprehensive approach to deploying the Incident Management System in a production environment. Adjust the configurations based on your specific infrastructure and requirements.