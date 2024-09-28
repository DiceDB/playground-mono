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

Steps to clone and run:
```
$ git clone https://github.com/dicedb/playground-mono
$ cd playground-mono
$ go run main.go
```
