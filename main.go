package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/coreos/go-systemd/v22/dbus"
)

type Service struct {
	Name             string
	Triggered        bool
	ConsecutiveFails int
}

func (s *Service) check(dbusConn *dbus.Conn) {
	prop, err := dbusConn.GetUnitProperties(s.Name)
	if err != nil {
		log.Fatal(err)
	}
	subState := prop["SubState"]
	if subState == "running" {
		if s.ConsecutiveFails > 0 {
			log.Printf("%s is running again after %d failed checks\n", s.Name, s.ConsecutiveFails)

		} else {
			log.Printf("%s is running\n", s.Name)
		}
		s.ConsecutiveFails = 0
		s.Triggered = false
	} else {
		s.ConsecutiveFails++
		log.Printf("%s is not running. %d consecutive failed check\n", s.Name, s.ConsecutiveFails)
	}
}

type arrayFlags []string

func (i *arrayFlags) String() string {
	return "my string representation"
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

type AlertServiceInterface interface {
	CreateIncident(string, string) error
}

func main() {
	var serviceNames arrayFlags
	flag.Var(&serviceNames, "service", "Name of the SystemD Services to monitor")
	durationPtr := flag.String("sleep", "1m", "Check interval")
	gracePeriodPtr := flag.String("grace-period", "", "Time before the first check")
	thresholdPtr := flag.Int("threshold", 1, "Amount of failed checks before alerting")
	customerNamePtr := flag.String("customer-name", "", "Name of the customer (Required)")
	alertingToolPtr := flag.String("alerting-tool", "", "Choose 'pagerduty' or 'victorops' for alerting (Required)")
	serviceKeyPtr := flag.String("service-key", "", "Service Key for the PagerDuty service")
	restIDPtr := flag.String("rest-id", "", "REST ID for VictorOps")
	restKeyPtr := flag.String("rest-key", "", "REST Key for VictorOps")
	flag.Parse()

	// Check that customer-name was set
	if *customerNamePtr == "" {
		fmt.Println("'-customer-name' is required")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Check that we have at least one service name
	if len(serviceNames) == 0 {
		fmt.Println("Please provide at least one service name")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Check that the correct API keys where used depending on the choosen alerting tool
	var alertService AlertServiceInterface
	switch *alertingToolPtr {
	case "pagerduty":
		if *serviceKeyPtr == "" {
			fmt.Println("Pageduty need the '-service-key' flag")
			flag.PrintDefaults()
			os.Exit(1)
		}
		alertService = PagerDuty{serviceKey: *serviceKeyPtr}
	case "victorops":
		if *restIDPtr == "" || *restKeyPtr == "" {
			fmt.Println("VictorOps need '-rest-key' and '-rest-id'")
			flag.PrintDefaults()
			os.Exit(1)
		}
		alertService = VictorOps{restID: *restIDPtr, restKey: *restKeyPtr}
	default:
		fmt.Printf("Unknown monitoring tool %q\n", *&alertingToolPtr)
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Grace period
	if *gracePeriodPtr != "" {
		log.Printf("Waiting %s before first check", *gracePeriodPtr)
		graceDuration, _ := time.ParseDuration(*gracePeriodPtr)
		time.Sleep(graceDuration)
	}

	dbusConn, err := dbus.NewSystemConnection()
	if err != nil {
		log.Fatal(err)
	}

	d, _ := time.ParseDuration(*durationPtr)

	var services []*Service
	for _, name := range serviceNames {
		services = append(services, &Service{Name: name, ConsecutiveFails: 0, Triggered: false})
	}

	for {
		for _, srv := range services {
			srv.check(dbusConn)
			if srv.ConsecutiveFails >= *thresholdPtr && !srv.Triggered {
				log.Printf("Service %q reached the threshold (%d). Creating an incident!\n", srv.Name, *thresholdPtr)
				err := alertService.CreateIncident(srv.Name, *customerNamePtr)
				if err != nil {
					fmt.Printf("Failed to create incident: %s\n", err.Error())
				} else {
					srv.Triggered = true
				}
			}
		}

		log.Printf("Sleeping for %s\n", *durationPtr)
		time.Sleep(d)
	}
}
