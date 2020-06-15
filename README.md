goMicroServiceStat
==================

goMicroServiceStat is a Go MQTT client to report status of a host machine (like a node running micro services) to an MQTT broker / topic.

current reported information
----------------------------

- online status
- host info aobut OS and disks
- current memory and swap status

This information is currently published to a local mqtt broker instance to the topic "monitoring/{hostname}". This can easily be modified in main.go (and in future via cmd line parameters).

usage
-----

- Clone this repository
- run: `go build`
- run: `./goMicroServiceStat`

easy testing
------------

- install docker to run the test environment
- run local mqtt broker: `docker run -it -p 1883:1883 -p 9001:9001 -v /mosquitto/data -v /mosquitto/log eclipse-mosquitto
`
- install the MQTT explorer to easily see published results. On Ubuntu run `sudo snap install mqtt-explorer`