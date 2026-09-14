$ErrorActionPreference = "Stop"
$Base = "http://localhost:8080"
$Session = New-Object Microsoft.PowerShell.Commands.WebRequestSession

Write-Host "1) Health"
Invoke-RestMethod -Uri "$Base/healthz" -WebSession $Session | ConvertTo-Json -Depth 5

Write-Host "2) Readiness"
Invoke-RestMethod -Uri "$Base/readyz" -WebSession $Session | ConvertTo-Json -Depth 5

Write-Host "3) Admin login"
$LoginBody = @{ email = "admin@example.com"; password = "ChangeMe123!" } | ConvertTo-Json
$Login = Invoke-RestMethod -Method Post -Uri "$Base/api/v1/auth/login" -ContentType "application/json" -Body $LoginBody -WebSession $Session
$Token = $Login.data.accessToken
if (-not $Token) { throw "No access token returned" }

Write-Host "4) Current user"
Invoke-RestMethod -Uri "$Base/api/v1/users/me" -Headers @{ Authorization = "Bearer $Token" } -WebSession $Session | ConvertTo-Json -Depth 5

Write-Host "5) Admin list users"
Invoke-RestMethod -Uri "$Base/api/v1/admin/users?page=1&limit=20&sortBy=createdAt&sortOrder=desc" -Headers @{ Authorization = "Bearer $Token" } -WebSession $Session | ConvertTo-Json -Depth 5

Write-Host "6) Refresh token rotation"
$Refresh = Invoke-RestMethod -Method Post -Uri "$Base/api/v1/auth/refresh" -ContentType "application/json" -Body '{}' -WebSession $Session
if (-not $Refresh.data.accessToken) { throw "Refresh did not return a new access token" }

Write-Host "Integration smoke flow passed"
