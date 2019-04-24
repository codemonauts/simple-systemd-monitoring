# simple-systemd-monitoring

Simple tool to check if a systemd service is running and otherwise create an alert via PagerDuty or VictorOps

## Get API credentials for PagerDuty
Create a new service with the integration type `APIv2` at the [PagerDuty website](https://codemonauts.pagerduty.com/services). This should give you an `integration-key`.

## Get API credentials for VictorOps
Login to VictorOps and visit `Settings -> API` which shows your `api-id`. To get a new `api-key` simly press the `+ New Key` button at the end.

##  Installation
 * Download and compile the code with `go get github.com/codemonauts/simple-systemd-monitoring` or use the [precompiled binaries](https://github.com/codemonauts/simple-systemd-monitoring/releases).
 *  Run the code with
```
./simple-systemd-monitoring -customerName Somebody -integration-key foobarbaz42 -service-name nginx.service -duration 30s
```
or use the provided `simple-systemd-monitoring.service` file.


