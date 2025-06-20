# convert-img: 
#  	convert images/I_Left1.jpg images/I_Left1_converted.jpg


build: 
	go build -o bin/simple/fingerprint sample/*.go

build-static:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/static/fingerprint sample/*.go
