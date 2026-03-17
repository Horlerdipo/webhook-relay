package enums

type HttpMethod string

const (
	Get    HttpMethod = "GET"
	Post   HttpMethod = "POST"
	Put    HttpMethod = "PUT"
	Patch  HttpMethod = "PATCH"
	Delete HttpMethod = "DELETE"
)
