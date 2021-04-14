# sample-service
['[![Build Status](https://jenkins.edgexfoundry.org/view/EdgeX%20Foundry%20Project/job/edgexfoundry/job/sample-service/job/master/badge/icon)](https://jenkins.edgexfoundry.org/view/EdgeX%20Foundry%20Project/job/edgexfoundry/job/sample-service/job/master/)', '[![Code Coverage](https://codecov.io/gh/edgexfoundry/sample-service/branch/master/graph/badge.svg?token=QrtB3XMRUl)](https://codecov.io/gh/edgexfoundry/sample-service)', '[![Go Report Card](https://goreportcard.com/badge/github.com/edgexfoundry/sample-service)](https://goreportcard.com/report/github.com/edgexfoundry/sample-service)', '[![GitHub Tag)](https://img.shields.io/github/v/tag/edgexfoundry/sample-service?include_prereleases&sort=semver&label=latest)](https://github.com/edgexfoundry/sample-service/tags)', '[![GitHub License](https://img.shields.io/github/license/edgexfoundry/sample-service)](https://choosealicense.com/licenses/apache-2.0/)', '![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/edgexfoundry/sample-service)', '[![GitHub Pull Requests](https://img.shields.io/github/issues-pr-raw/edgexfoundry/sample-service)](https://github.com/edgexfoundry/sample-service/pulls)', '[![GitHub Contributors](https://img.shields.io/github/contributors/edgexfoundry/sample-service)](https://github.com/edgexfoundry/sample-service/contributors)', '[![GitHub Commit Activity](https://img.shields.io/github/commit-activity/m/edgexfoundry/sample-service)](https://github.com/edgexfoundry/sample-service/commits)']

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
