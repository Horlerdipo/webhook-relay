package enums

import "fmt"

type VerificationType string

const (
	None           VerificationType = "none"
	RequestSigning VerificationType = "request_signing"
	StaticToken    VerificationType = "static_token"
)

func (h VerificationType) MarshalBinary() ([]byte, error) {
	return []byte(h), nil
}

func ParseVerificationType(s string) (VerificationType, error) {
	switch s {
	case "none":
		return None, nil
	case "request_signing":
		return RequestSigning, nil
	case "static_token":
		return StaticToken, nil
	default:
		return "", fmt.Errorf("invalid verification type: %s", s)
	}
}

type VerificationKeyLocation string

const (
	NoLocation VerificationKeyLocation = "no_location"
	Header     VerificationKeyLocation = "header"
	Body       VerificationKeyLocation = "body"
)

func (h VerificationKeyLocation) MarshalBinary() ([]byte, error) {
	return []byte(h), nil
}

func ParseVerificationKeyLocation(s string) (VerificationKeyLocation, error) {
	switch s {
	case "no_location":
		return NoLocation, nil
	case "header":
		return Header, nil
	case "body":
		return Body, nil
	default:
		return "", fmt.Errorf("invalid verification key location: %s", s)
	}
}
