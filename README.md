# Leaseplan-Bot

The Leaseplan bot is a Leaseplanabocar to Telegram Bridge which allows users to login to their Leaseplan account and recieve notifications when there are any changes in the available offerings.

The Bot is mainly build around the following interfaces:
- [leaseplanabocarexporter](https://github.com/khase/leaseplanabocarexporter)
- [telegram-bot-api](https://github.com/go-telegram-bot-api/telegram-bot-api)

If you wan't to support me you can always follow the [contribution guide](#contribution) or value my time by [buying me a Pizza](https://www.paypal.me/KenHasenbank) (figuratively)

## Basic functionality
To be done. 

## Contribution

If you wan't to improve the bot feel free to create any Pull-Requests or point out Bugs, problems or feature Requests via a Github issue.

## Setup hosting/developing

### The blue pill

The easiest way to deploy the bot on your own Infrastructure is to use the Docker-Image built automagically by this repository:
[Leaseplan-Bot on Docker Hub](https://hub.docker.com/repository/docker/khase/leaseplan-bot)

The following composefile will start the latest image of the bot with all features enabled:

```yaml
version: "3.9"
services:
  bot:
    image: khase/leaseplan-bot:latest
    ports:
      - 2112:2112
    volumes:
      - ./data/leaseplan-bot.userdata:/opt/leaseplan-bot.userdata
      - ./data/cache:/opt/cache
    command: start -t <Telegram-Bot-Token>
```

The `.userdata` mount is necessary for the bot to remember all it's connected clients and their leaseplan login information. Without it the bot won't remember any users across restarts

The `cache` mount is used to persist any leaseplan data across restarts of the bot. The cache makes it possible to determine any changes between the last scrape of the old instance and the first scrape of the new instance.

The port mapping `2112:2112` is used to make phe prometheus metrics endpoint reachable through the host.

Even though none of the mentioned settings are truely necessary i strongly reccoment to use them to provide the complete experience.

The bare minimum would be to provide the bot a telegram api token:
```sh
docker run khase/leaseplan-bot:latest start -t <Telegram-Bot-Token>
``` 
This docker command is sufficient to test the bots functionality

### The red pill

The red pill is ment for developers to be able to implement their own features and build the bot from scratch.

To be able to build the bot from scratch you either need a local installation of the Golang tools or a running docker engine.

#### Run localy

Install instructions for the golang tools can be found at the official go page https://go.dev/doc/install

##### Run tests
```sh
go test ./...
``` 

##### Run
```sh
go run ./leaseplan-bot.go start -t <Telegram-Bot-Token>
``` 

##### Build
```sh
go build -o build/leaseplan-bot ./leaseplan-bot.go
``` 

#### Build Docker Image

To build the binary inside a Docker container no special configuration is needed.
The dockerfile is composed of a two-step build where it uses the official `golang` image to build the binary and copies it in a low wheigt `archlinux`.

```sh
docker build -t leaseplan-bot .
``` 

To run the built docker container you follow the instructions above in the [`blue pill section`](#the-blue-pill)

