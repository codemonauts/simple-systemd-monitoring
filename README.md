# simple-systemd-monitoring

Simple tool to check if a systemd service is running and otherwise create an PagerDuty event

##  Usage
 * Create a new service with the integration type `APIv2` at [PagerDuty website](https://codemonauts.pagerduty.com/services)
 * Download and compile the code with `go get github.com/codemonauts/simple-systemd-monitoring`
 *  Run the code with
```
./worker-monitoring -customerName Somebody -integration-key foobarbaz42 -service-name nginx.service -duration 30s
```


