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

func createPagerdutyEvent(serviceKey string, customer string) error {
	description := fmt.Sprintf("The worker from %s failed", customer)
	event := pagerduty.Event{
		Type:        "trigger",
		ServiceKey:  serviceKey,
		Description: description,
	}
	_, err := pagerduty.CreateEvent(event)
	if err != nil {
		return err
	}

	return nil
}

type VictoropsIncident struct {
	Behaviour   string `json:"message_type"`
	Description string `json:"entity_display_name"`
}

func createVictoropsEvent(restID string, restKey string, customer string) error {
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
	_, err = client.Do(req)
	if err != nil {
		return err
	}

	return nil
}

type arrayFlags []string

func (i *arrayFlags) String() string {
	return "my string representation"
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func checkService(dbusConn *dbus.Conn, name string) bool {
	prop, err := dbusConn.GetUnitProperties(name)
	if err != nil {
		log.Fatal(err)
	}
	subState := prop["SubState"]
	if subState == "running" {
		log.Printf("%s is running\n", name)
		return true
	} else {
		log.Printf("%s is not running\n", name)
		return false
	}
}

func main() {
	var serviceNames arrayFlags
	flag.Var(&serviceNames, "service", "Name of the SystemD Services to monitor")
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

	// Check that we have at least one service name
	if len(serviceNames) == 0 {
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

	d, _ := time.ParseDuration(*durationPtr)
	var createErr error

	for {
		for _, name := range serviceNames {
			if !checkService(dbusConn, name) {
				switch *alertingToolPtr {
				case "pagerduty":
					log.Println("Service is not running! Creating an alert with Pagerduty")
					createErr = createPagerdutyEvent(*integrationKeyPtr, *customerNamePtr)
				case "victorops":
					log.Println("Service is not running! Creating an alert with VictorOps")
					createErr = createVictoropsEvent(*restIDPtr, *restKeyPtr, *customerNamePtr)
				}
			}
			if createErr != nil {
				fmt.Errorf("Failed to create incident: %s\n", createErr)
			} else {
				log.Printf("SUcessfully created an incident. Sleeping for 15m")
				time.Sleep(time.Minute * 15)
			}
		}

		log.Printf("Sleeping for %d\n", d)
		time.Sleep(d)
	}
}
