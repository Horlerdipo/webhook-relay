package enums

type VerificationType string

const (
	RequestSigning VerificationType = "RequestSigning"
	StaticToken    VerificationType = "StaticToken"
)

type VerificationKeyLocation string

const (
	Header VerificationKeyLocation = "Header"
	Body   VerificationKeyLocation = "Body"
)
