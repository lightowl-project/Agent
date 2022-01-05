Expand-Archive ..\telegraf-1.21.1_windows_amd64.zip 'C:\Program Files\' -Force

$SERVER_ADDR = $args[0]
$API_KEY = $args[1]

Copy-Item -Force -Recurse '.\lightowl\etc\telegraf\telegraf.d' 'C:\Program Files\telegraf-1.21.1\'
Copy-Item -Force -Recurse '.\lightowl\etc\lightowl' 'C:\Program Files\'

$DATA = @{
    os = "Windows"
    hostname = $env:computername
    tags = @()
    plugins = @{}
}

$HEADERS = @{
    api_key = $API_KEY
    "Content-Type" = "application/json"
}

$URI = "https://$SERVER_ADDR/api/v1/agents/join"

[System.Net.ServicePointManager]::ServerCertificateValidationCallback = { $true }
Invoke-RestMethod -Uri $URI -Method Post -Headers $HEADERS -Body ($DATA|ConvertTo-Json) -OutFile C:\Windows\Temp\lightowl.zip

Set-Location C:\Windows\Temp\
Expand-Archive .\lightowl.zip C:\Windows\Temp\
Set-Location C:\Windows\Temp\lightowl

Copy-Item .\.env 'C:\Program Files\lightowl\' -Force
Copy-Item .\ca.pem 'C:\Program Files\lightowl\ssl\' -Force

Copy-Item .\lightowl.conf 'C:\Program Files\telegraf-1.21.1\telegraf.d\' -Force
Copy-Item .\telegraf.conf 'C:\Program Files\telegraf-1.21.1\telegraf.conf' -Force

Remove-Item C:\Windows\Temp\lightowl.zip
Remove-Item C:\Windows\Temp\lightowl