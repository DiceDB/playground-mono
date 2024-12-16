# DiceDB Playground Mono

[DiceDB Playground](https://playground.dicedb.io/) is an interactive platform designed to let users experiment with [DiceDB](https://github.com/dicedb/dice/) commands in a live environment. Allows users to search and play with various DiceDB commands in real-time. This repository hosts backend service implementation of the Playground.

### Setting up for development and contributions

To run playground-mono for local development or running from source, you will need

1. [Golang](https://go.dev/)
2. Any of the below supported platform environment:
    1. [Linux based environment](https://en.wikipedia.org/wiki/Comparison_of_Linux_distributions)
    2. [OSX (Darwin) based environment](https://en.wikipedia.org/wiki/MacOS)
    3. WSL under Windows
3. Install GoLangCI

```bash
sudo su
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b /bin v1.60.1
```

### Setup

```sh
git clone https://github.com/dicedb/playground-mono
cd playground-mono
cp .env.sample .env
```

> The default values of .env files works just fine, but
> feel free to tweak them as per your environment.

### Run

#### Pre-requisite

Playground-mono requires two running instances of DiceDB one on port `7379` and
other on port `7380`.

To setup and run DiceDB locally, please refer to the
[README.md file of DiceDB/dice](https://github.com/DiceDB/dice/blob/master/README.md).

> If you do not want to run two instances, due to hardware limitations, you can run one
> and change the .env file and update the appropriate values

#### Start the server

Once you have DiceDB running, run the following command and that will start
the playground-mono server on port `8080`.

```bash
make run
```

## Setting up a Production instance

```
bash setup.sh
```

## How to contribute

The code contribution guidelines are published at [CONTRIBUTING.md](CONTRIBUTING.md); please read them before you start making any changes. This would allow us to have a consistent standard of coding practices and developer experience.

Contributors can join the [Discord Server](https://discord.gg/6r8uXWtXh7) for quick collaboration.
