set dotenv-load

init:
  cp .env.example .env

rebuild:
    go clean
    go build

run:
    clear
    source .env
    go run main.go

send-message:
    curl -X POST -H \
        "Content-Type: application/json" -d \
        '{"sender": "just-cli-user", "message": "hello, rollup"}' \
        localhost:8080/message

docker-reset:
    ./docker-compose/reset.sh

docker-build:
    docker-compose -f docker-compose/local.yaml build

docker-run:
    clear
    open http://localhost:3000
    if [ ! -f ./docker-compose/.data/cometbft/priv_validator_state.json ]; then \
        ./docker-compose/reset.sh; \
    fi
    docker-compose -f docker-compose/local.yaml up
