package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/PagerDuty/go-pagerduty"
	"github.com/coreos/go-systemd/dbus"
)

func createEvent(serviceKey string, customer string) {
	description := fmt.Sprintf("The worker from %s failed", customer)
	event := pagerduty.Event{
		Type:        "trigger",
		ServiceKey:  serviceKey,
		Description: description,
	}
	resp, err := pagerduty.CreateEvent(event)
	if err != nil {
		log.Println(resp)
		log.Fatalln("ERROR:", err)
	}

	log.Println("Incident key:", resp.IncidentKey)
}

func main() {
	serviceNamePtr := flag.String("service-name", "worker.service", "Name of the SystemD Service to monitor")
	durationPtr := flag.String("sleep", "1m", "Time to sleep between checking of the service is running")
	customerNamePtr := flag.String("customer-name", "", "Name of the customer (Required")
	integrationKeyPtr := flag.String("integration-key", "", "Integration Key for the PagerDuty service (Required")
	flag.Parse()

	if *integrationKeyPtr == "" || *customerNamePtr == "" {
		flag.PrintDefaults()
	}

	dbusConn, err := dbus.NewSystemConnection()
	if err != nil {
		log.Fatal(err)
	}

	for {
		prop, err := dbusConn.GetUnitProperties(*serviceNamePtr)
		if err != nil {
			log.Fatal(err)
		}
		subState := prop["SubState"]
		if subState != "running" {
			createEvent(*integrationKeyPtr, *customerNamePtr)
		}

		fmt.Printf("Service is running. Sleeping for %s", *durationPtr)
		d, _ := time.ParseDuration(*durationPtr)
		time.Sleep(d)
	}
}
