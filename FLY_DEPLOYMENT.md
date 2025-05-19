# Deploying to Fly.io

This document outlines the steps to deploy the Modern Band Backend to Fly.io with SQLite persistence.

## Prerequisites

1. Install the Fly.io CLI
   ```
   # Windows (PowerShell)
   iwr https://fly.io/install.ps1 -useb | iex
   ```

2. Add the Fly.io CLI to your path
   ```
   $env:Path += ";$HOME\.fly\bin"
   ```

3. Login to Fly.io
   ```
   flyctl auth login
   ```

4. Add a credit card to your Fly.io account (required even for free tier)
   Visit: https://fly.io/dashboard/personal/billing

## Deployment Steps

1. **Create the Fly.io application**
   ```
   flyctl apps create modern-band-api
   ```

2. **Create a persistent volume for SQLite data**
   ```
   flyctl volumes create modern_band_data --size 1 --region bom
   ```

3. **Deploy the application**
   ```
   flyctl deploy
   ```

4. **Check that the app is running**
   ```
   flyctl status
   ```

## Using the Deployment Script

Instead of running the commands individually, you can use the provided script:

```
./deploy.ps1
```

## Database Backup

To backup your SQLite database from Fly.io:

```
./backup.ps1
```

This will create a timestamped backup in the `./backups` directory.

## Troubleshooting

1. **View logs**
   ```
   flyctl logs
   ```

2. **SSH into the application**
   ```
   flyctl ssh console
   ```

3. **Check SQLite database**
   ```
   flyctl ssh console -C "sqlite3 /app/data/app.db '.tables'"
   ```

## Scaling & Resource Management

Your app is deployed on the Fly.io free tier which includes:
- 3 shared-cpu-1x 256MB VMs
- 3GB of persistent volume storage
- 160GB outbound data transfer

To upgrade or adjust resources, visit: https://fly.io/dashboard/personal/billing 