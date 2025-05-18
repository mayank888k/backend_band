#!/usr/bin/env pwsh
# Backup script for SQLite database on Fly.io

$date = Get-Date -Format "yyyy-MM-dd-HHmm"
$backupDir = "./backups"
$backupFile = "$backupDir/app-db-$date.sqlite"

# Create the backup directory if it doesn't exist
if (-not (Test-Path $backupDir)) {
    New-Item -ItemType Directory -Path $backupDir
}

Write-Host "Creating backup of SQLite database from Fly.io..."

# Connect to the app and create a backup of the SQLite database
flyctl ssh console -C "sqlite3 /app/data/app.db '.backup /tmp/app-backup.db'"

# Copy the backup file from the app to the local machine
flyctl ssh sftp get /tmp/app-backup.db $backupFile

Write-Host "Backup completed successfully! File saved to $backupFile" 