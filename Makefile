sqlc:
	cd repo/db && sqlc generate

swagger:
	swag init --output server/api/swagger -g server/api/swagger/main.go

isza: sqlc swagger
	CGO_ENABLED=0 go build -o isza main.go

run-server: isza
	godotenv -f .env ./isza \
		--port 8000 \
		--var-dir var \
		--dev-logging \
		--insta-working-dir ./instagram

clean:
	rm -f isza