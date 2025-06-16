# Project ObservabilityServer

The server for the Observability project

## Getting Started

These instructions will get you a copy of the project up and running on your local machine for development and testing purposes. See deployment for notes on how to deploy the project on a live system.

#### To run a local database use the following command:

```bash
docker run -e POSTGRES_DB=observability -e POSTGRES_PASSWORD=postgres -v ~/.docker/volumes/observe_db_data:/var/lib/postgresql/data -p 127.0.0.1:5432:5432/tcp -d postgres
```

#### To migrate your local datbase run this command:

> ensure that you have a '.env' file with the variables used in the follwing command

```bash
source .env

migrate -path ./migrations -database postgres://$POSTGRES_USER:$POSTGRES_PASSWORD@$POSTGRES_HOST:$POSTGRES_PORT/$POSTGRES_DB?sslmode=disable up
```

#### To inspect your database you can use a docker image like adminer:

```bash
docker run -p 127.0.0.1:8000:8080/tcp -d adminer
```

## MakeFile

Run build make command with tests

```bash
make all
```

Build the application

```bash
make build
```

Run the application

```bash
make run
```

Create DB container

```bash
make docker-run
```

Shutdown DB Container

```bash
make docker-down
```

DB Integrations Test:

```bash
make itest
```

Live reload the application:

```bash
make watch
```

Run the test suite:

```bash
make test
```

Clean up binary from the last build:

```bash
make clean
```
