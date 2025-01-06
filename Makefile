ENV_FILE=.env

run:
	@echo "Running application..."
	@set -a; source $(ENV_FILE); go run cmd/main.go