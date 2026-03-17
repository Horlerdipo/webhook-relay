package routeregistrar

import (
	"context"
	"github.com/horlerdipo/webhook-relay/internal/datastore"
	"github.com/horlerdipo/webhook-relay/internal/enums"
)

type Route struct {
	HttpMethod              enums.HttpMethod
	Name                    string
	Identifier              string
	VerificationType        enums.VerificationType
	VerificationKeyLocation enums.VerificationKeyLocation
	VerificationKeyName     string
	VerificationToken       string
	Active                  bool
}

type Destination struct {
	HttpMethod        enums.HttpMethod
	Identifier        string
	Active            bool
	Url               string
	VerificationToken string
}

type RouteRegistrar interface {
	AddRoute(ctx *context.Context, route Route) error
	RemoveRoute(ctx *context.Context, routeId string) error
	FetchRoutes(ctx *context.Context) ([]Route, error)
	fetchRoute(ctx *context.Context, routeId string) (Route, error)
	AddDestination(ctx *context.Context, routeId string, destination Destination) error
	RemoveDestination(ctx *context.Context, routeId string, destinationId string) error
}

type DefaultRouteRegistrar struct {
	datastore *datastore.Store
}

//func (ds *DefaultRouteRegistrar) AddRoute(ctx *context.Context, route Route) error {
//	//generate uuid
//	//add route
//	//return nil
//}
