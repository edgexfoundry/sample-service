# sample-service

## Not for production use
Sample service is a basic hello-world application that is used by the EdgeX DevOps working group for build automation / validation of Jenkins Pipelines.

The base functionality is derived from the 'device-random' service. This is a working service that will generate a random int8 value when queried.

Get an edgeX docker-compose from [developer-scripts](https://github.com/edgexfoundry/developer-scripts/tree/master/releases)


To start the sample-service:
```
cd /sample-service
make clean
make docker
docker-compose up sample-service
```

Confirm the service `edgexfoundry/docker-sample-service-go` is running:
```
docker-compose ps
```

To access the API endpoints, in a webbrowser, CURL or POSTMAN,
```
http://localhost:48082/api/v1/device/name/sample-service01
```
The command endpoint for 'GenerateRandomValue_Int8' will be displayed and can be directly accessed. Take note that `edge-core-command` should be replaced by 'localhost' if you are running locally and the '<unique_ID>' fields are unique to your instance of EdgeX.
```
http://localhost:48082/api/v1/device/<unique_ID>/command/<unique_ID>
```

Useful References:

[edgex-examples](https://github.com/edgexfoundry/edgex-examples)

[device-random](https://github.com/edgexfoundry/device-random)
