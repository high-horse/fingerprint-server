# convert-img: 
#  	convert images/I_Left1.jpg images/I_Left1_converted.jpg

SERVICE_NAME=fingerprint


build: 
	go build -o bin/simple/fingerprint sample/*.go

build-static:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/static/fingerprint sample/*.go


# New Target for server side use
stop:
	@echo "Stopping ${SERVICE_NAME} service..."
	sudo systemctl stop ${SERVICE_NAME}.service

start:
	@echo "Starting ${SERVICE_NAME} service..."
	sudo systemctl start ${SERVICE_NAME}.service

restart:
	@echo "Restarting ${SERVICE_NAME} service..."
	sudo systemctl restart ${SERVICE_NAME}.service

status:
	@echo "Checking status of ${SERVICE_NAME} service..."
	sudo systemctl status ${SERVICE_NAME}.service


deploy:
	@echo "Stopping Pre existing service..."
	make stop
	@echo "Pulling from git..."
	git pull origin main
	chmod +x bin/static/fingerprint
	@echo "restarting the service..."
	make start

test:
	@echo "Running tests..."
	@curl -sSf http://127.0.0.1:9090/health || echo "Health check failed"
