package enums

import "fmt"

type HttpMethod string

const (
	Get    HttpMethod = "Get"
	Post   HttpMethod = "Post"
	Put    HttpMethod = "Put"
	Patch  HttpMethod = "Patch"
	Delete HttpMethod = "Delete"
)

func (h HttpMethod) MarshalBinary() ([]byte, error) {
	return []byte(h), nil
}

func ParseHttpMethod(s string) (HttpMethod, error) {
	switch s {
	case "Get":
		return Get, nil
	case "Post":
		return Post, nil
	case "Put":
		return Put, nil
	case "Patch":
		return Patch, nil
	case "Delete":
		return Delete, nil
	default:
		return "", fmt.Errorf("invalid http method: %s", s)
	}
}
