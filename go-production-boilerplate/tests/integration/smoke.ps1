$Base = "http://localhost:8080"
Invoke-RestMethod "$Base/healthz"
Invoke-RestMethod "$Base/readyz"
Write-Host "Health smoke test passed"
