package main

import (
	"fmt"

	"github.com/PagerDuty/go-pagerduty"
)

type PagerDuty struct {
	serviceKey string
}

func (pd PagerDuty) CreateIncident(service string, customer string) error {
	description := fmt.Sprintf("The service %q from %s failed", service, customer)
	event := pagerduty.Event{
		Type:        "trigger",
		ServiceKey:  pd.serviceKey,
		Description: description,
	}
	_, err := pagerduty.CreateEvent(event)
	if err != nil {
		return err
	}
	return nil
}
