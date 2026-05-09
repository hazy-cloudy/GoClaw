$ErrorActionPreference = "Stop"

function Get-PetWebSocketUrl {
    $tokenResp = Invoke-RestMethod -Uri "http://127.0.0.1:18790/pet/token"
    $wsUrl = [string]$tokenResp.ws_url
    if ([string]::IsNullOrWhiteSpace($wsUrl)) {
        throw "pet token endpoint did not return ws_url"
    }
    if ($wsUrl -match "\?") {
        return "${wsUrl}&sessionId=apifox-debug"
    }
    return "${wsUrl}?sessionId=apifox-debug"
}

function Invoke-PetWsAction {
    param(
        [Parameter(Mandatory = $true)]
        [string]$WebSocketUrl,
        [Parameter(Mandatory = $true)]
        [string]$Action
    )

    Add-Type -AssemblyName System.Net.Http

    $ws = [System.Net.WebSockets.ClientWebSocket]::new()
    $uri = [Uri]$WebSocketUrl
    $cts = [System.Threading.CancellationTokenSource]::new()
    $cts.CancelAfter([TimeSpan]::FromSeconds(10))

    try {
        $ws.ConnectAsync($uri, $cts.Token).GetAwaiter().GetResult()

        $payload = @{
            action     = $Action
            request_id = "debug-$([DateTimeOffset]::UtcNow.ToUnixTimeMilliseconds())"
            data       = @{}
        } | ConvertTo-Json -Compress

        $bytes = [System.Text.Encoding]::UTF8.GetBytes($payload)
        $segment = [System.ArraySegment[byte]]::new($bytes)
        $ws.SendAsync(
            $segment,
            [System.Net.WebSockets.WebSocketMessageType]::Text,
            $true,
            $cts.Token
        ).GetAwaiter().GetResult()

        $sawActionResponse = $false
        $targetPushTypes = @("weekly_report_ready", "progress_nudge")

        while (-not $cts.IsCancellationRequested) {
            $buffer = New-Object byte[] 8192
            $receiveSegment = [System.ArraySegment[byte]]::new($buffer)
            $result = $ws.ReceiveAsync($receiveSegment, $cts.Token).GetAwaiter().GetResult()
            if ($result.MessageType -eq [System.Net.WebSockets.WebSocketMessageType]::Close) {
                break
            }

            $text = [System.Text.Encoding]::UTF8.GetString($buffer, 0, $result.Count)
            if ([string]::IsNullOrWhiteSpace($text)) {
                continue
            }

            Write-Host $text

            $json = $null
            try {
                $json = $text | ConvertFrom-Json
            }
            catch {
                continue
            }

            if ($json.action -eq $Action) {
                $sawActionResponse = $true
            }

            if ($json.type -eq "push" -and $targetPushTypes -contains [string]$json.push_type) {
                continue
            }

            if ($sawActionResponse) {
                break
            }
        }
    }
    finally {
        if ($ws.State -eq [System.Net.WebSockets.WebSocketState]::Open) {
            $ws.CloseOutputAsync(
                [System.Net.WebSockets.WebSocketCloseStatus]::NormalClosure,
                "done",
                [System.Threading.CancellationToken]::None
            ).GetAwaiter().GetResult()
        }
        $ws.Dispose()
        $cts.Dispose()
    }
}

$wsUrl = Get-PetWebSocketUrl

Write-Host "Testing weekly report..." -ForegroundColor Cyan
Invoke-PetWsAction -WebSocketUrl $wsUrl -Action "debug_generate_weekly_report"
Write-Host ""
Write-Host "Testing progress nudge..." -ForegroundColor Cyan
Invoke-PetWsAction -WebSocketUrl $wsUrl -Action "debug_generate_progress_nudge"
