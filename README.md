# PERIL

## About The Project

A Go application that uses RabbitMQ for message queuing and publish/subscribe communication, with server and client components that interact through RabbitMQ exchanges and queues. From the Boot.dev "Learn Pub/Sub" course, per the module path in `go.mod` (`github.com/bootdotdev/learn-pub-sub-starter`); described as an exploratory project in the original project description.

### Built With

- Go 1.22
- RabbitMQ (amqp091-go)
- Docker

## Getting Started

### Prerequisites

- Go 1.22+
- Docker

### Running the RabbitMQ broker

```sh
./rabbit.sh start
```

This builds and runs a RabbitMQ 3.13 container with the management UI (port 15672) and the STOMP plugin enabled. `./rabbit.sh stop` stops it and `./rabbit.sh logs` follows its logs.

### Running the server

```sh
go run ./cmd/server
```

`multiserver.sh <number-of-instances>` starts several server instances at once for testing.

## License

No license file is present in this repository.

## Contact

Dyrits - https://github.com/Dyrits
