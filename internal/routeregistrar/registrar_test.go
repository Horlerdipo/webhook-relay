package routeregistrar

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/horlerdipo/webhook-relay/internal/datastore"
	"github.com/horlerdipo/webhook-relay/internal/enums"
	"github.com/horlerdipo/webhook-relay/internal/models"
	"github.com/horlerdipo/webhook-relay/internal/testutils"
	"math/rand"
	"testing"
)

func TestDefaultRouteRegistrar_AddRoute(t *testing.T) {
	redisStore, redisInstance, cleanup := testutils.NewTestRedisStore(t)
	defer cleanup()
	ctx := context.Background()
	registrar := NewDefaultRouteRegistrar(redisStore)
	route := testutils.NewRouteModel()
	routeId, err := registrar.AddRoute(ctx, route)

	if err != nil {
		t.Fatalf("AddRoute() error = %v", err)
	}

	if routeId != route.Identifier {
		t.Fatalf("wrong route id, got %s, want %s", routeId, route.Identifier)
	}

	//confirm that the details was saved into redis
	isMember, _ := redisInstance.SIsMember(datastore.RoutesKey, route.Identifier)
	if !isMember {
		t.Fatalf("expected route '%s' to be a member of the '%s' set, not member", route.Identifier, datastore.RoutesKey)
	}

	verificationToken := redisInstance.HGet(fmt.Sprintf("%s%s", datastore.RouteKey, route.Identifier), "verification_token")
	if verificationToken != route.VerificationToken {
		t.Fatalf("expected verification token to be '%s', got '%s'", route.VerificationToken, verificationToken)
	}

	//it should return error that the route already exists
	routeId, err = registrar.AddRoute(ctx, route)
	if err == nil {
		t.Fatal("expected AddRoute to return error, got nil")
	} else {
		errString := fmt.Sprintf("Route with ID %s already exists", route.Identifier)
		if err.Error() != errString {
			t.Fatalf("expected error to be '%s', got '%s'", errString, err.Error())
		}
	}

	//it should return error when verification type is None and key location is not no_location
	route.VerificationKeyLocation = enums.Header
	route.Identifier = uuid.NewString()
	routeId, err = registrar.AddRoute(ctx, route)

	if err == nil {
		t.Fatalf("AddRoute() should retun an error, got nil")
	} else {
		errString := "verification key location must be `no location` if verification type is `None`"
		if err.Error() != errString {
			t.Fatalf("expected error to be %v, got %v", errString, err.Error())
		}
	}
}

func TestDefaultRouteRegistrar_RemoveRoute(t *testing.T) {
	redisStore, redisInstance, cleanup := testutils.NewTestRedisStore(t)
	defer cleanup()

	ctx := context.Background()
	registrar := NewDefaultRouteRegistrar(redisStore)
	route := testutils.NewRouteModel()
	_, err := redisStore.AddRoute(ctx, route)
	if err != nil {
		return
	}

	err = registrar.RemoveRoute(ctx, route.Identifier)
	if err != nil {
		t.Fatalf("expected err to be nil, got %s", err)
	}

	//confirm on Redis that the details were deleted
	isMember, _ := redisInstance.SIsMember(datastore.RoutesKey, route.Identifier)
	if isMember {
		t.Fatalf("expected route '%s' to no longer be a member of the '%s' set, still member", route.Identifier, datastore.RoutesKey)
	}

	verificationToken := redisInstance.HGet(fmt.Sprintf("%s%s", datastore.RouteKey, route.Identifier), "verification_token")
	if verificationToken != "" {
		t.Fatalf("expected verification token to be empty got '%s'", verificationToken)
	}

	//try removing route that does not exist
	err = registrar.RemoveRoute(ctx, route.Identifier)
	if err == nil {
		t.Fatal("expected error, got nil")
	} else {
		errString := "route does not exist"
		if err.Error() != errString {
			t.Fatalf("expected error to be '%s', got '%s'", errString, err.Error())
		}
	}
}

func TestDefaultRouteRegistrar_FetchRoutes(t *testing.T) {
	redisStore, _, cleanup := testutils.NewTestRedisStore(t)
	defer cleanup()

	ctx := context.Background()
	registrar := NewDefaultRouteRegistrar(redisStore)

	randomNumber := rand.Intn(10)
	var routes []models.Route
	for i := 1; i <= randomNumber; i++ {
		route := testutils.NewRouteModel()
		_, err := redisStore.AddRoute(ctx, route)
		if err != nil {
			t.Fatalf("AddRoute() error = %v", err)
		}
		routes = append(routes, route)
	}

	fetchedRoutes, err := registrar.FetchRoutes(ctx, false)
	if err != nil {
		t.Fatalf("FetchRoutes() error = %v", err)
	}

	if len(fetchedRoutes) != len(routes) {
		t.Fatalf("expected %d routes, got %d", len(routes), len(fetchedRoutes))
	}

	for _, singleRoute := range fetchedRoutes {
		found := false
		for _, route := range routes {
			if singleRoute.Identifier == route.Identifier {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("unable to find route %s in the list of fetched routes", singleRoute.Identifier)
		}
	}
}

func TestDefaultRouteRegistrar_FetchRoute(t *testing.T) {
	redisStore, _, cleanup := testutils.NewTestRedisStore(t)
	defer cleanup()

	ctx := context.Background()
	registrar := NewDefaultRouteRegistrar(redisStore)
	route := testutils.NewRouteModel()

	_, err := redisStore.AddRoute(ctx, route)
	if err != nil {
		t.Fatalf("AddRoute() error = %v", err)
	}

	fetchedRoute, err := registrar.FetchRoute(ctx, route.Identifier, false)
	if err != nil {
		t.Fatalf("FetchRoute() error = %v", err)
	}

	if fetchedRoute.Identifier != route.Identifier {
		t.Fatalf("expected route identifier to be %s, got %s", route.Identifier, fetchedRoute.Identifier)
	}

	//remove route and confirm that it returns error
	err = redisStore.RemoveRoute(ctx, route.Identifier)
	if err != nil {
		t.Fatalf("RemoveRoute() error = %v", err)
	}

	fetchedRoute, err = registrar.FetchRoute(ctx, route.Identifier, false)
	if err == nil {
		t.Fatal("expected error, got nil")
	} else {
		errString := "route does not exist"
		if err.Error() != errString {
			t.Fatalf("expected error to be '%s', got '%s'", errString, err.Error())
		}
	}
}

func TestDefaultRouteRegistrar_AddDestination(t *testing.T) {
	redisStore, _, cleanup := testutils.NewTestRedisStore(t)
	defer cleanup()

	ctx := context.Background()
	registrar := NewDefaultRouteRegistrar(redisStore)
	route := testutils.NewRouteModel()

	_, err := redisStore.AddRoute(ctx, route)
	if err != nil {
		t.Fatalf("AddRoute() error = %v", err)
	}

	destination := testutils.NewDestinationModel(route.Identifier)
	destinationId, err := registrar.AddDestination(ctx, route.Identifier, destination)
	if err != nil {
		t.Fatalf("AddDestination() error = %v", err)
	}

	if destinationId != destination.Identifier {
		t.Fatalf("expected destination identifier to be %s, got %s", destination.Identifier, destinationId)
	}

	//returns error if route does not exist
	destinationId, err = registrar.AddDestination(ctx, uuid.NewString(), destination)
	if err == nil {
		t.Fatal("expected error, got nil")
	} else {
		errString := "route does not exist"
		if err.Error() != errString {
			t.Fatalf("expected error to be '%s', got '%s'", errString, err.Error())
		}
	}

	destinationId, err = registrar.AddDestination(ctx, route.Identifier, destination)
	if err == nil {
		t.Fatal("expected error, got nil")
	} else {
		errString := "destination already exists"
		if err.Error() != errString {
			t.Fatalf("expected error to be '%s', got '%s'", errString, err.Error())
		}
	}
}

func TestDefaultRouteRegistrar_RemoveDestination(t *testing.T) {
	redisStore, redisInstance, cleanup := testutils.NewTestRedisStore(t)
	defer cleanup()

	ctx := context.Background()
	registrar := NewDefaultRouteRegistrar(redisStore)
	route := testutils.NewRouteModel()
	_, err := redisStore.AddRoute(ctx, route)
	if err != nil {
		t.Fatalf("AddRoute() error = %v", err)
	}

	destination := testutils.NewDestinationModel(route.Identifier)

	destinationsKey := fmt.Sprintf("%s:%s", datastore.DestinationsKey, route.Identifier)
	destinationKey := fmt.Sprintf("%s%s", datastore.DestinationKey, destination.Identifier)

	_, err = redisStore.AddDestination(ctx, route.Identifier, destination)
	if err != nil {
		t.Fatalf("AddDestination() error = %v", err)
	}

	err = registrar.RemoveDestination(ctx, route.Identifier, destination.Identifier)
	if err != nil {
		t.Fatalf("expected err to be nil, got %s", err)
	}

	//confirm on Redis that the details were deleted
	isMember, _ := redisInstance.SIsMember(destinationsKey, destination.Identifier)
	if isMember {
		t.Fatalf("expected route '%s' to no longer be a member of the '%s' set, still member", route.Identifier, datastore.RoutesKey)
	}

	identifier := redisInstance.HGet(fmt.Sprintf("%s%s", destinationKey, route.Identifier), "identifier")
	if identifier != "" {
		t.Fatalf("expected identifier to be empty got '%s'", identifier)
	}

	//try removing destination that does not exist
	err = registrar.RemoveDestination(ctx, route.Identifier, destination.Identifier)
	if err == nil {
		t.Fatal("expected error, got nil")
	} else {
		errString := "destination does not exist"
		if err.Error() != errString {
			t.Fatalf("expected error to be '%s', got '%s'", errString, err.Error())
		}
	}
}

func TestDefaultRouteRegistrar_FetchRouteDestinations(t *testing.T) {
	redisStore, _, cleanup := testutils.NewTestRedisStore(t)
	defer cleanup()

	ctx := context.Background()
	registrar := NewDefaultRouteRegistrar(redisStore)

	route := testutils.NewRouteModel()
	_, err := registrar.AddRoute(ctx, route)
	if err != nil {
		t.Fatalf("AddRoute() error = %v", err)
	}

	randomNumber := rand.Intn(10)
	var destinations []models.Destination
	for i := 1; i <= randomNumber; i++ {
		destination := testutils.NewDestinationModel(route.Identifier)
		_, err := redisStore.AddDestination(ctx, route.Identifier, destination)
		if err != nil {
			t.Fatalf("AddDestination() error = %v", err)
		}
		destinations = append(destinations, destination)
	}

	fetchedDestinations, err := registrar.FetchRouteDestinations(ctx, route.Identifier)
	if err != nil {
		t.Fatalf("FetchRouteDestinations() error = %v", err)
	}

	if len(fetchedDestinations) != len(destinations) {
		t.Fatalf("expected %d routes, got %d", len(fetchedDestinations), len(destinations))
	}

	for _, singleDestination := range fetchedDestinations {
		found := false
		for _, destination := range destinations {
			if destination.Identifier == singleDestination.Identifier {
				found = true
				break
			}
		}
		if !found {
			t.Log(fetchedDestinations)
			t.Log(destinations)
			t.Fatalf("unable to find destination %s in the list of fetched destinations", singleDestination.Identifier)
		}
	}
}

func TestDefaultRouteRegistrar_FetchDestinationDetails(t *testing.T) {
	redisStore, _, cleanup := testutils.NewTestRedisStore(t)
	defer cleanup()

	ctx := context.Background()
	registrar := NewDefaultRouteRegistrar(redisStore)
	route := testutils.NewRouteModel()

	_, err := redisStore.AddRoute(ctx, route)
	if err != nil {
		t.Fatalf("AddRoute() error = %v", err)
	}

	destination := testutils.NewDestinationModel(route.Identifier)
	_, err = redisStore.AddDestination(ctx, route.Identifier, destination)
	if err != nil {
		t.Fatalf("AddDestination() error = %v", err)
	}

	fetchedDestination, err := registrar.FetchDestinationDetails(ctx, route.Identifier, destination.Identifier)
	if err != nil {
		t.Fatalf("FetchDestinationDetails() error = %v", err)
	}

	if fetchedDestination.Identifier != destination.Identifier {
		t.Fatalf("expected route identifier to be %s, got %s", destination.Identifier, fetchedDestination.Identifier)
	}

	//remove destination and confirm that it returns error
	err = redisStore.RemoveDestination(ctx, route.Identifier, destination.Identifier)
	if err != nil {
		t.Fatalf("RemoveDestination() error = %v", err)
	}

	fetchedDestination, err = registrar.FetchDestinationDetails(ctx, route.Identifier, destination.Identifier)
	if err == nil {
		t.Fatal("expected error, got nil")
	} else {
		errString := "destination does not exist"
		if err.Error() != errString {
			t.Fatalf("expected error to be '%s', got '%s'", errString, err.Error())
		}
	}
}
