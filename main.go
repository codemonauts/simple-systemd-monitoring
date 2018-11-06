package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/PagerDuty/go-pagerduty"
	"github.com/coreos/go-systemd/dbus"
)

func createPagerdutyEvent(serviceKey string, customer string) {
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

type VictoropsIncident struct {
	Behaviour   string `json:"message_type"`
	Description string `json:"entity_display_name"`
}

func createVictoropsEvent(restID string, restKey string, customer string) {
	description := fmt.Sprintf("The worker from %s failed", customer)
	i := VictoropsIncident{
		Behaviour:   "CRITICAL",
		Description: description,
	}
	restEndpoint := fmt.Sprintf("https://alert.victorops.com/integrations/generic/%s/alert/%s/%s", restID, restKey, customer)

	jsonStr, _ := json.Marshal(i)
	req, err := http.NewRequest("POST", restEndpoint, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
}

func main() {
	serviceNamePtr := flag.String("service-name", "worker.service", "Name of the SystemD Service to monitor")
	durationPtr := flag.String("sleep", "1m", "Time to sleep between checking of the service is running")
	customerNamePtr := flag.String("customer-name", "", "Name of the customer (Required)")
	alertingToolPtr := flag.String("alerting-tool", "", "Choose 'pagerduty' or 'victorops' for alerting (Required)")
	integrationKeyPtr := flag.String("integration-key", "", "Integration Key for the PagerDuty service")
	restIDPtr := flag.String("rest-id", "", "REST ID for VictorOps")
	restKeyPtr := flag.String("rest-key", "", "REST Key for VictorOps")
	flag.Parse()

	// Check that customer-name was set
	if *customerNamePtr == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Check that the correct API keys where used depending on the choosen alerting tool
	switch *alertingToolPtr {
	case "pagerduty":
		if *integrationKeyPtr == "" {
			flag.PrintDefaults()
			os.Exit(1)
		}
	case "victorops":
		if *restIDPtr == "" || *restKeyPtr == "" {
			flag.PrintDefaults()
			os.Exit(1)
		}
	default:
		flag.PrintDefaults()
		os.Exit(1)
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
			switch *alertingToolPtr {
			case "pagerduty":
				log.Println("Service is not running! Creating an alert with Pagerduty")
				createPagerdutyEvent(*integrationKeyPtr, *customerNamePtr)
			case "victorops":
				log.Println("Service is not running! Creating an alert with VictorOps")
				createVictoropsEvent(*restIDPtr, *restKeyPtr, *customerNamePtr)
			}
		}

		fmt.Printf("Service is running. Sleeping for %s\n", *durationPtr)
		d, _ := time.ParseDuration(*durationPtr)
		time.Sleep(d)
	}
}
