go-config
====

A key-value configuration management service written in Go


### Requirements

- Docker - to run the service
- NodeJS - to test the service


### Run with
```
cd app
docker-compose up
```

### Tests setup
```
npm install -g newman
```

### Run the tests
```
cd tests
newman run go-config.postman_collection.json
```

### Features

CRUD service for configurations in the format
```
{
  "id": "some-id",
  "name": "Configuration for Foo",
  "value": "This is the value for configuration Foo"
}
```

- GET /someid - get configuration for id "someid"
- POST /someid {config} - store configuration with id "someid"
- PUT /someid {config} - override configuration with id "someid"
- DELETE /someid - delete configuration
- GET / - retrieve all configurations

### Known limitations


- The service is currently using a single connection to Redis DB. Connections pooling may be beneficial for performances / concurrency.