set dotenv-load

init:
  cp .env.example .env

run:
    clear
    go run main.go

send-message:
    curl -X POST -H \
        "Content-Type: application/json" -d \
        '{"sender": "itamar", "message": "hello, rollup", "priority": 1}' \
        localhost:8080/message
