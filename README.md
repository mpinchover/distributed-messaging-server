# To run the service

## Create the docket network

```
docker network create external-example
```

## Get the databases up and running first

```
cd dependencies
docker compose up --build
```

## Get the service up and running

```
cd messaging-service
docker compose up --build
```

## Run integration tests

```
cd messaging-service/integration-tests
go test
```

# Misc.

```
 docker build --tag messaging-server . --no-cache
```

```
 docker run -d -p 5002:9090 messaging-server
```

Dockerfile uses host.docker.internal instead of localhost to connect to the localhost of the host machine.

Build all the images first

```
docker compose up -d --build
```

connect to locally running dockerized mysql

```
mysql -h 127.0.0.1 -P 3308 -u root -p
```

Restart just one container in docker compose

```
docker-compose restart msgserver
```

Start a single service in docker compose

```
docker-compose up -d --no-deps --build <service_name>

```

## TODO's

### P1

- Add an API key/auth in some way

  - once the client connects they should get back a JWT. For all socket related req, just use JWT.
  - use redis to track API keys

- allow members, rooms to have a stringified text field that can track whatever the user wants.
- update tests to run in go routines to mimic high, concurrent volumes

### P2

- separate out channel and client events
- allow someone to be invited to the chat
- Add a "typing" attribute/event to the message
- Run processMessage in go routine and inform the client if a message fails. This will let the message be routed directly to the client
- run UpdateMessageToSeen in go routine
- on the server side, mark the messages as seen so the client doesn't have to

### P3

- controltower should be parent of
  - realtime controller (sockets)
  - message controller (sync flows for messages)
- LeaveRoom should also save messages that someone has left the chat
- app should show if someone screenshotted
- add in a user permissions table to link to member table
