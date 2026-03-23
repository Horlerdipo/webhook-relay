package routeregistrar

import (
	"context"
	"errors"
	"github.com/horlerdipo/webhook-relay/internal/datastore"
	"github.com/horlerdipo/webhook-relay/internal/enums"
	"github.com/horlerdipo/webhook-relay/internal/models"
)

type RouteRegistrar interface {
	AddRoute(ctx context.Context, route models.Route) error
	RemoveRoute(ctx context.Context, routeId string) error
	FetchRoutes(ctx context.Context) ([]models.Route, error)
	FetchRoute(ctx context.Context, routeId string) (models.Route, error)
	AddDestination(ctx context.Context, routeId string, destination models.Destination) error
	RemoveDestination(ctx context.Context, routeId string, destinationId string) error
	FetchRouteDestinations(ctx context.Context, routeId string) ([]models.Destination, error)
}

type DefaultRouteRegistrar struct {
	Datastore datastore.Store
}

func (drr *DefaultRouteRegistrar) AddRoute(ctx context.Context, route models.Route) error {
	if route.VerificationType == enums.None && route.VerificationKeyLocation != enums.NoLocation {
		return errors.New("verification key location must be `no location` if verification type is `None`")
	}
	return drr.Datastore.AddRoute(ctx, route)
}

func (drr *DefaultRouteRegistrar) RemoveRoute(ctx context.Context, routeId string) error {
	//check if route exists
	routeExists, err := drr.Datastore.CheckRouteExistence(ctx, routeId)
	if err != nil {
		return err
	}

	if !routeExists {
		return errors.New("route does not exist")
	}

	//if it does, remove
	return drr.Datastore.RemoveRoute(ctx, routeId)
}

func (drr *DefaultRouteRegistrar) FetchRoutes(ctx context.Context, withDestinations bool) ([]models.Route, error) {
	//check if route exists
	return drr.Datastore.FetchRoutes(ctx, withDestinations)
}

func (drr *DefaultRouteRegistrar) FetchRoute(ctx context.Context, routeId string, withDestinations bool) (models.Route, error) {
	routeExists, err := drr.Datastore.CheckRouteExistence(ctx, routeId)
	if err != nil {
		return models.Route{}, err
	}

	if !routeExists {
		return models.Route{}, errors.New("route does not exist")
	}
	return drr.Datastore.FetchRoute(ctx, routeId, withDestinations)
}

func (drr *DefaultRouteRegistrar) AddDestination(ctx context.Context, routeId string, destination models.Destination) error {
	//check if route exists
	routeExists, err := drr.Datastore.CheckRouteExistence(ctx, routeId)
	if err != nil {
		return err
	}

	if !routeExists {
		return errors.New("route does not exist")
	}

	//check if destination does not exist
	destinationExists, _ := drr.Datastore.CheckDestinationExistence(ctx, routeId, destination.Identifier)
	if destinationExists {
		return errors.New("destination already exists")
	}

	//create destination
	err = drr.Datastore.AddDestination(ctx, routeId, destination)
	if err != nil {
		return err
	}

	return nil
}
