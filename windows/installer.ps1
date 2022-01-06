Expand-Archive .\telegraf.zip 'C:\Program Files\' -Force

$SERVER_ADDR = $args[0]
$API_KEY = $args[1]

$TELEGRAF_CONFIG = 'C:\Program Files\telegraf-1.21.1\telegraf.conf'
$TELEGRAF_LIGHTOWL_CONFIG = "C:\Program Files\telegraf-1.21.1\telegraf.d\lightowl.conf"
$LIGHTOWL_BINARY = "C:\Program Files\lightowl\lightowl.exe"

Copy-Item -Force -Recurse '.\etc\telegraf\telegraf.d' 'C:\Program Files\telegraf-1.21.1\'
Copy-Item -Force -Recurse '.\etc\lightowl' 'C:\Program Files\'

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

Copy-Item .\lightowl.conf $TELEGRAF_LIGHTOWL_CONFIG -Force
Copy-Item .\telegraf.conf $TELEGRAF_CONFIG -Force

Set-Location 'C:\Program Files\telegraf-1.21.1\'
Remove-Item C:\Windows\Temp\lightowl.zip
Remove-Item C:\Windows\Temp\lightowl -Recurse

.\telegraf.exe --service install --config $TELEGRAF_CONFIG --config-directory $TELEGRAF_LIGHTOWL_CONFIG
.\telegraf.exe --service start

schtasks.exe /create /tn LightOwl /sc minute /tr cmd.exe /c $LIGHTOWL_BINARY