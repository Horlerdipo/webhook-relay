package models

import "github.com/horlerdipo/webhook-relay/internal/enums"

type Route struct {
	HttpMethod              enums.HttpMethod              `json:"http_method,omitempty" redis:"http_method"`
	Name                    string                        `json:"name,omitempty" redis:"name"`
	Identifier              string                        `json:"identifier,omitempty" redis:"identifier"`
	VerificationType        enums.VerificationType        `json:"verification_type,omitempty" redis:"verification_type"`
	VerificationKeyLocation enums.VerificationKeyLocation `json:"verification_key_location,omitempty" redis:"verification_key_location"`
	VerificationKeyName     string                        `json:"verification_key_name,omitempty" redis:"verification_key_name"`
	VerificationToken       string                        `json:"verification_token,omitempty" redis:"verification_token"`
	Active                  bool                          `json:"active,omitempty" redis:"active"`
	Destinations            []Destination
}
