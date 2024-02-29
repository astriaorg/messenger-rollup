set dotenv-load

init:
  cp .env.example .env

rebuild:
    go clean
    go build

run:
    clear
    go run main.go

docker-reset:
    ./docker-compose/reset.sh

docker-build:
    docker-compose -f docker-compose/local.yaml build

docker-run:
    clear
    open http://localhost:3000
    docker-compose -f docker-compose/local.yaml up

send-message:
    curl -v -X POST -H \
        "Content-Type: application/json" -d \
        '{"sender": "itamar", "message": "hello, rollup", "priority": 1}' \
        localhost:8080/message
