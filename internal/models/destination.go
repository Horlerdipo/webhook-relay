package models

import "github.com/horlerdipo/webhook-relay/internal/enums"

type Destination struct {
	HttpMethod        enums.HttpMethod `json:"http_method,omitempty" redis:"http_method"`
	Identifier        string           `json:"identifier,omitempty" redis:"identifier"`
	Active            bool             `json:"active,omitempty" redis:"active"`
	Url               string           `json:"url,omitempty" redis:"url"`
	VerificationToken string           `json:"verification_token,omitempty" redis:"verification_token"`
	RouteIdentifier   string           `json:"route_identifier,omitempty" redis:"route_identifier"`
}
