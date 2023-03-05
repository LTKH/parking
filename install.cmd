cd /d %~dp0
parking64.exe --service stop
parking64.exe --service uninstall
parking64.exe --service install
parking64.exe --service start
start http://localhost:8000
pause