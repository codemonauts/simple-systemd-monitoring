package main

import (
	"fmt"

	"github.com/imroc/req"
)

type VictorOps struct {
	restID  string
	restKey string
}

type VictoropsIncident struct {
	Behaviour   string `json:"message_type"`
	Description string `json:"entity_display_name"`
}

func (vo VictorOps) CreateIncident(service string, customer string) error {
	description := fmt.Sprintf("The service %q from %s failed", service, customer)
	incident := VictoropsIncident{
		Behaviour:   "CRITICAL",
		Description: description,
	}

	restEndpoint := fmt.Sprintf("https://alert.victorops.com/integrations/generic/%s/alert/%s/%s", vo.restID, vo.restKey, customer)
	header := req.Header{"Content-Type": "application/json"}
	_, err := req.Post(restEndpoint, req.BodyJSON(&incident), header)
	if err != nil {
		return err
	}

	return nil
}
