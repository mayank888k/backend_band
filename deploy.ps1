#!/usr/bin/env pwsh
# Deployment script for Fly.io

# Create the app if it doesn't exist
flyctl apps create modern-band-api --org personal

# Create a volume for persistent SQLite data
flyctl volumes create modern_band_data --size 1 --region sin

# Deploy the application
flyctl deploy

Write-Host "Deployment complete! Your app should be running at https://modern-band-api.fly.dev" 