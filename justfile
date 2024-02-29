set dotenv-load

init:
  cp .env.example .env

rebuild:
    go clean
    go build

run:
    clear
    go run main.go

send-message:
    curl -v -X POST -H \
        "Content-Type: application/json" -d \
        '{"sender": "itamar", "message": "hello, rollup", "priority": 1}' \
        localhost:8080/message
