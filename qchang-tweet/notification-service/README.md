# notification-service

notification-service is a subscriber service for working with tweet service. 
This service will trigger when user add comment to some post in tweet it push message sending via rabbitmq.
After that the notification-service handle message and send notification data to post owner by use websocket channel. 

## How to run

### Start Rabbitmq

```bash
docker-compose up -d
```

## Run the app

```bash
go run cmd/main.go
```
