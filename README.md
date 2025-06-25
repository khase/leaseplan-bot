# Leaseplan-Bot
![Docker Pulls](https://img.shields.io/docker/pulls/khase/leaseplan-bot)

The Leaseplan bot is a Leaseplanabocar to Telegram Bridge which allows users to login to their Leaseplan account and recieve notifications when there are any changes in the available offerings.

The Bot is mainly build around the following interfaces:

- [leaseplanabocarexporter](https://github.com/khase/leaseplanabocarexporter)
- [telegram-bot-api](https://github.com/go-telegram-bot-api/telegram-bot-api)

If you wan't to support me you can always follow the [contribution guide](#contribution) or value my time by [buying me a Pizza](https://www.paypal.me/KenHasenbank) (figuratively)

## End User documentation

The bot provides a Telegram based User-Interface.
The following commands are currently available:

Command                                             | Short description
----------------------------------------------------|------------------------------------------------------
[start](#start)                                     | creates a new internal bot-user
[whoami](#whoami)                                   | returns all data the bot knows about you
[resume](#resume)                                   | activates change notifications
[pause](#pause)                                     | deactivates change notifications
[login](#login)                                     | log-in to leaseplan using username/password
[settoken](#settoken)                               | log-in to leaseplan using your access token
[~~connect~~](#connect)                             | ~~subscribe to change notifications of another user~~
[throttle](#throttle)                               | sets a minumum delay between updates
[ignoreDetails](#ignoreDetails)                     | controls weather or not your updates contain detail informations about cars
[ignoreRemoved](#ignoreRemoved)                     | controls weather or not your updates contain informations about removed cars
[setsummarymessageformat](#setsummarymessageformat) | updates personal summary message format
[setdetailmessageformat](#setdetailmessageformat)   | updates personal detail message format
[test](#test)                                       | returns a test message
[filter](#filter)                                   | sets a filter for your update message

### start

The `start` command is automatically sent when first approaching to the bot and hitting the "Start" button in the bottom of the chat window.
To signal a successful registration the bot will greet you 😉

### whoami

The command will return all data the bot currently knows about you.
Including for example your leaseplan token as well as your currently set summary and detail message format.

It just returns a yaml encoded version of the [user struct](lpbot/config/user.go) associated with your user.

### resume

Sending a `resume` will set activate the watcher responsible for your leaseplan account which will notify you for any changes in your leaseplan offer.

**⚠ Known Issue:** The command currently is a little buggy where it won't actually start your watcher. It will just mark your watcher as "to be started" and the bot needs to be restarted for this change to take affect. Contact your Bot-Admin and ask him to restart the bot.

### pause

Deactivates your watcher and thus you won't get any more change updates.
Effectively makes the bot to "shut up".

### login

This command is the first of two ways to login to the leaseplan api.
The command has to be formatted as follows:

```command
/login <leaseplan-email> <leaseplan-password>
```

After a successful login the bot will internally save a token used for retrieving leaseplan data in your name and delete the credentials message in the history.

### settoken

This command is the second of two ways to login to the leaseplan api.
The command has to be formatted as follows:

```command
/settoken <leaseplan-token>
```

After a successful login the bot will internally save a token used for retrieving leaseplan data in your name and delete the credentials message in the history.

### connect

This command was originally planed as a way to "piggibag" to another users updates.
Unfortunately the command is not yet implemented. 😔

### throttle

This command throttles your messages to be limited to 1 update every n Minutes.
Note: One update can consist of multiple messages beeing sent since one message has a limited capacity

```command
/throttle <throttle rate in minutes>
```

### ignoreDetails

This command sets a flag for your user that controlls weater or not your updates should contain detail informations about cars.
Using the command without any parameters enables the feature effectively skipping generation of detail messages.

To disable the feature add any of the following as an argument: `0`, `f`, `F`, `FALSE`, `false`, `False`

```command
/ignoreDetails
/ignoreDetails 0
```


### ignoreRemoved

This command sets a flag for your user that controlls weater or not your updates should contain detail informations about removed cars.
Using the command without any parameters enables the feature effectively skipping generation of detail messages for removed cars.

To disable the feature add any of the following as an argument: `0`, `f`, `F`, `FALSE`, `false`, `False`

```command
/ignoreRemoved
/ignoreRemoved 0
```



### setsummarymessageformat

Using this command you can overwrite your personal summary message format. Internally the bot uses the `html/template` engine to validate your format. A detailed documentation can be found at the [official package documentation](https://pkg.go.dev/html/template#Template) (this documentation is quite "tecky" but i did not find somethin more beginner friendly yet 🙁).

The passed root object is the current [dataframe](lpbot/config/dataFrame.go) providing you with the complete `current` car list, `previous` car list as well as it's changes represented as `added` and `removed`.

Example (current default message formats can be found in the [user struct](lpbot/config/user.go)):

```template
{{ len .Previous }} -> {{ len .Current }} (+{{ len .Added }}, -{{ len .Removed }})
```

Which will render as following:

```text
Änderungen: 113 -> 112 (+0, -1)
```

Additionally the engine is extended using the [Masterminds/sprig package](https://github.com/Masterminds/sprig) and some custom functions:

function  | parameter | description
----------|-----------|------------------------------------------------------------------------------------------------------------------------------------------------------------
portalUrl | dto.Item  | returns the portal url pointing to the given car formatted with its offer name
taxPrice  | dto.Item  | returns the tax price for the given car -> 1% of the net cost for normal ICE cars, 0.5% for PHEV and BEV, 0.25% for BEV cars with a net cost lower than 60k
netCost   | dto.Item  | returns an approximate total net cost for the car based on the individual salery waiver and the taxPrice.
italic    | string    | wraps the string in underscores so that Telegram will render it in an italic font
bold      | string    | wraps the string in asteriks so that Telegram will render it in an bold font

### setdetailmessageformat

Using this command you can overwrite your personal detail message format. Internally the bot uses the `html/template` engine to validate your format. A detailed documentation can be found at the [official package documentation](https://pkg.go.dev/html/template#Template) (this documentation is quite "tecky" but i did not find somethin more beginner friendly yet 🙁).

The passed root object is [dto.Item](https://github.com/khase/leaseplanabocarexporter/blob/master/dto/item.go) which contains all known data about a single car offer.
The most interesting Data (e.g. car model, net price or engine type) can be found in the property [RentalObject](https://github.com/khase/leaseplanabocarexporter/blob/master/dto/rental_object.go)

Example (current default message formats can be found in the [user struct](lpbot/config/user.go)):

```template
{{ portalUrl . }}
  PS: {{ .RentalObject.PowerHP }}, Antrieb: {{ .RentalObject.KindOfFuel }}
  BLP: {{ .RentalObject.PriceProducer1 }}€, BGV: {{.SalaryWaiver}}€, Netto: ~{{ round ( netCost . ) 2 }}€
  Verfügbar: {{.RentalObject.DateRegistration.Format \"02.01.2006\"}}
```

Which will render as following:

```text
Some manufacturers car 
  PS: 1337, Antrieb: Plug-in-Hybrid
  BLP: 59999€, BGV: 333€, Netto: ~318.6€  
  Verfügbar: 24.12.2022
```

Additionally the engine is extended using the [Masterminds/sprig package](https://github.com/Masterminds/sprig) and some custom functions:

function  | parameter | description
----------|-----------|------------------------------------------------------------------------------------------------------------------------------------------------------------
portalUrl | dto.Item  | returns the portal url pointing to the given car formatted with its offer name
taxPrice  | dto.Item  | returns the tax price for the given car -> 1% of the net cost for normal ICE cars, 0.5% for PHEV and BEV, 0.25% for BEV cars with a net cost lower than 60k
netCost   | dto.Item  | returns an approximate total net cost for the car based on the individual salery waiver and the taxPrice.
italic    | string    | wraps the string in underscores so that Telegram will render it in an italic font
bold      | string    | wraps the string in asteriks so that Telegram will render it in an bold font

### test

The command can be used to test your set message formats.
It uses the last dataframe collected by the bot in your name and formates it using your currently set formats. this will effectively replay the last dataframe with your current config.

When the last dataframe does not contain any changes you will only get a summary message. In this scenario you can force the command to add any number of changes to the frame by specifying an argument (2 added and 2 removed for this example):

```command
/test 2
```

### filter

The filter command can be used to filter out specific car items.
The filter list can be manipulated with the three subcommands `list`, `add` and `remove`

You can define multiple filters which will be combined in an `or` fashion. (if one filter fails that item will be removed from your list)

For evaluation the filters the `html/template` engine is used (much like in the templating section)
Therefore you can access all properties the same way as in the `setdetailmessageformat` section.

short recap:
The passed root object is [dto.Item](https://github.com/khase/leaseplanabocarexporter/blob/master/dto/item.go) which contains all known data about a single car offer.
The most interesting Data (e.g. car model, net price or engine type) can be found in the property [RentalObject](https://github.com/khase/leaseplanabocarexporter/blob/master/dto/rental_object.go)

But in contrast you don't need the most outer curly braces `{{}}`.

The Template has to evaluate to a boolean expression (e.g. `true`, `false`, `0`, `1`)
Filters that do not evaluate in a known boolean value or do fail in any other way are ignored and don't affect the result list.

Comparisons are built using a function like structure `operator arg1 arg2` and for quickstart the following operators are supported:

Operator | Description
---------|--------------------------------------------
eq       | Returns the boolean truth of arg1 == arg2
ne       | Returns the boolean truth of arg1 != arg2
lt       | Returns the boolean truth of arg1 < arg2
le       | Returns the boolean truth of arg1 <= arg2
gt       | Returns the boolean truth of arg1 > arg2
ge       | Returns the boolean truth of arg1 >= arg2
and      | Returns the boolean truth of arg1 && arg2
or       | Returns the boolean truth of arg1 || arg2

A detailed documentation can be found at the [official package documentation](https://pkg.go.dev/html/template#Template)

#### Examples

##### List allyour currently active filters
```
/filter list
```

##### Add Filter to ignore all cars from VOLVO
```filter
ne (.RentalObject.CarLabel | lower) "volvo"
```
We are here retrieving the CarLabel (which is used to indicate the manufacturer by leaseplan) and converting it to it's lowercase representation to make the condition case-insensitive.

To add this filter simply use the following command:
```
/filter add ne (.RentalObject.CarLabel | lower) "volvo"
```

##### Add Filter to ignore all cars with less than 300 HP
```filter
gt .RentalObject.PowerHP 300
```

To add this filter simply use the following command:
```
/filter add gt .RentalObject.PowerHP 300
```

##### Remove Filter to ignore all cars from VOLVO
```
/filter remove ne (.RentalObject.CarLabel | lower) "volvo"
```

## Contribution

If you wan't to improve the bot feel free to create any Pull-Requests or point out Bugs, problems or feature Requests via a Github issue.

## Setup hosting/developing

### The blue pill

The easiest way to deploy the bot on your own Infrastructure is to use the Docker-Image built automagically by this repository:
[Leaseplan-Bot on Docker Hub](https://hub.docker.com/repository/docker/khase/leaseplan-bot)

The following composefile will start the latest ``master`` image of the bot with all features enabled:

```yaml
version: "3.9"
services:
  bot:
    image: khase/leaseplan-bot:master
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
docker run khase/leaseplan-bot:master start -t <Telegram-Bot-Token>
```

This docker command is sufficient to test the bots functionality

### The red pill

The red pill is ment for developers to be able to implement their own features and build the bot from scratch.

To be able to build the bot from scratch you either need a local installation of the Golang tools or a running docker engine.

#### Run localy

Install instructions for the golang tools can be found at the official go page <https://go.dev/doc/install>

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
