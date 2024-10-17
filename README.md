DiceDB Playground Backend
===

DiceDB Playground is an interactive platform designed to let users experiment with [DiceDB](https://github.com/dicedb/dice/) commands in a live environment, similar to the Go Playground.
Allows users to search and play with various DiceDB commands in real-time.

This repository hosts backend service implementation of the Playground.

## How to contribute

The code contribution guidelines are published at [CONTRIBUTING.md](CONTRIBUTING.md); please read them before you start making any changes. This would allow us to have a consistent standard of coding practices and developer experience.

Contributors can join the [Discord Server](https://discord.gg/6r8uXWtXh7) for quick collaboration.

### Setting up this repository from source for development and contributions

To run playground-mono for local development or running from source, you will need

1. [Golang](https://go.dev/)
2. Any of the below supported platform environment:
    1. [Linux based environment](https://en.wikipedia.org/wiki/Comparison_of_Linux_distributions)
    2. [OSX (Darwin) based environment](https://en.wikipedia.org/wiki/MacOS)
    3. WSL under Windows
3. Install GoLangCI
```
$ sudo su
$ curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b /bin v1.60.1
```

### Steps to clone and run:
```sh
$ git clone https://github.com/dicedb/playground-mono
$ cd playground-mono
$ cp .env.sample .env
$ go run main.go
```

### Running the Project Using Docker

#### 1. Clone the repository:

```bash
git clone https://github.com/dicedb/playground-mono
```

#### 2. Navigate to the project directory:

```bash
cd playground-mono
```

#### 3. Copy the sample environment file:

```bash
cp .env.sample .env
```
This creates the `.env` file, which stores your environment variables. Make sure to update any necessary values inside this file before running the server.

#### 4. Start the application using Docker Compose:

```bash
docker compose up -d
```
This command will pull any necessary Docker images, build the containers, and run them in detached mode (`-d`).

#### 5. Verify the server is running:

Open your browser and go to:

```bash
http://localhost:8000/health
```
This endpoint should return a status indicating that the server is up and running.