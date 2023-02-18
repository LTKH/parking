sc stop Parking
sc create Parking binpath="c:\Parking\parking64.exe" start=auto
sc start Parking
start http://localhost:8000