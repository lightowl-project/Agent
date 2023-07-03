schtasks.exe /delete /tn LightOwl /f

Set-Location 'C:\Program Files\telegraf-1.21.1'
./telegraf.exe --service stop
./telegraf.exe --service uninstall

Set-Location C:\
Remove-Item -Force -Recurse 'C:\Program Files\telegraf-1.21.1'
Remove-Item -Force -Recurse 'C:\Program Files\lightowl'

Write-Host "LightOwl Agent successfully uninstalled"