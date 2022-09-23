claim:
	sudo chown -R $(whoami):$(id -g -n) *

start: claim
	docker-compose --env-file .env.dev up --build

start-background: claim
	docker-compose --env-file .env.dev up --build -d

db: claim
	docker-compose up postgres
