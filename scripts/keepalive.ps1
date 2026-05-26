param(
  [string]$Url = "https://your-render-service.onrender.com/api/v1/subscription-plans",
  [int]$IntervalSeconds = 600,
  [int]$TimeoutSeconds = 20
)

Write-Host "[keepalive] Starting ping loop"
Write-Host "[keepalive] Url=$Url Interval=${IntervalSeconds}s Timeout=${TimeoutSeconds}s"

while ($true) {
  $timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
  try {
    $response = Invoke-WebRequest -Uri $Url -Method GET -TimeoutSec $TimeoutSeconds -UseBasicParsing
    Write-Host "[$timestamp] OK status=$($response.StatusCode)"
  }
  catch {
    Write-Host "[$timestamp] ERROR $($_.Exception.Message)"
  }

  Start-Sleep -Seconds $IntervalSeconds
}
