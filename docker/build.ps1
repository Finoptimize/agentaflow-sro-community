#!/usr/bin/env pwsh
# PowerShell script to build all AgentaFlow Docker images

Write-Host "🐳 Building AgentaFlow Docker Images" -ForegroundColor Cyan
Write-Host "=====================================" -ForegroundColor Cyan
Write-Host ""

$ErrorActionPreference = "Stop"

# Navigate to project root
$projectRoot = Split-Path -Parent $PSScriptRoot
Set-Location $projectRoot

Write-Host "📍 Project root: $projectRoot" -ForegroundColor Yellow
Write-Host ""

# Build web-dashboard
Write-Host "🔨 Building web-dashboard image..." -ForegroundColor Green
docker build -f docker/Dockerfile.web-dashboard -t agentaflow-sro:web-dashboard .
if ($LASTEXITCODE -eq 0) {
    Write-Host "✅ web-dashboard build successful!" -ForegroundColor Green
} else {
    Write-Host "❌ web-dashboard build failed!" -ForegroundColor Red
    exit 1
}
Write-Host ""

# Build k8s-scheduler
Write-Host "🔨 Building k8s-scheduler image..." -ForegroundColor Green
docker build -f docker/Dockerfile.k8s-scheduler -t agentaflow-sro:k8s-scheduler .
if ($LASTEXITCODE -eq 0) {
    Write-Host "✅ k8s-scheduler build successful!" -ForegroundColor Green
} else {
    Write-Host "❌ k8s-scheduler build failed!" -ForegroundColor Red
    exit 1
}
Write-Host ""

# Build prometheus-demo
Write-Host "🔨 Building prometheus-demo image..." -ForegroundColor Green
docker build -f docker/Dockerfile.prometheus-demo -t agentaflow-sro:prometheus-demo .
if ($LASTEXITCODE -eq 0) {
    Write-Host "✅ prometheus-demo build successful!" -ForegroundColor Green
} else {
    Write-Host "❌ prometheus-demo build failed!" -ForegroundColor Red
    exit 1
}
Write-Host ""

# List built images
Write-Host "📦 Built Images:" -ForegroundColor Cyan
docker images | Select-String "agentaflow-sro"
Write-Host ""

Write-Host "✨ All images built successfully!" -ForegroundColor Green
Write-Host ""
Write-Host "🚀 Quick Start Commands:" -ForegroundColor Cyan
Write-Host "  docker run -p 9000:9000 -p 9001:9001 agentaflow-sro:web-dashboard" -ForegroundColor Yellow
Write-Host "  docker-compose up -d" -ForegroundColor Yellow
Write-Host ""
