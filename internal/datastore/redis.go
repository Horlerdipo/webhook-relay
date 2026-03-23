package datastore

import (
	"context"
	"errors"
	"fmt"
	"github.com/horlerdipo/webhook-relay/internal/enums"
	"github.com/horlerdipo/webhook-relay/internal/models"
	"github.com/redis/go-redis/v9"
	"time"
)

const RoutesKey = "routes:all"
const RouteKey = "routes:"
const DestinationsKey = "destinations:all"
const DestinationKey = "destinations"

type Store interface {
	Name() string
	Ping(ctx context.Context) error
	AddRoute(ctx context.Context, route models.Route) error
	CheckRouteExistence(ctx context.Context, routeId string) (bool, error)
	FetchRoutes(ctx context.Context, withDestination bool) ([]models.Route, error)
	FetchRoute(ctx context.Context, routeId string, withDestination bool) (models.Route, error)
	RemoveRoute(ctx context.Context, routeId string) error
	AddDestination(ctx context.Context, routeId string, destination models.Destination) error
	CheckDestinationExistence(ctx context.Context, routeId string, destinationId string) (bool, error)
	FetchDestinations(ctx context.Context, routeId string) ([]models.Destination, error)
	FetchDestination(ctx context.Context, routeId string, destinationId string) (models.Destination, error)
	RemoveDestination(ctx context.Context, destinationId string) error
}

type RedisStore struct {
	client *redis.Client
}

func (rs *RedisStore) Name() string {
	return "redis"
}

func (rs *RedisStore) Ping(ctx context.Context) error {
	err := rs.client.Ping(ctx).Err()
	if err != nil {
		return err
	}
	return nil
}

func (rs *RedisStore) AddRoute(ctx context.Context, route models.Route) (string, error) {
	//add route ID to Set
	res, err := rs.sAdd(ctx, RoutesKey, route.Identifier)
	if err != nil {
		return "", err
	}

	if res == 0 {
		return "", errors.New(fmt.Sprintf("Route with ID %s already exists", route.Identifier))
	}

	//add route details to Hash
	res, err = rs.hSet(ctx, fmt.Sprintf("%s%s", RouteKey, route.Identifier), route)
	if err != nil {
		_, _ = rs.sRem(ctx, RoutesKey, route.Identifier)
		return "", err
	}

	if res == 0 {
		_, _ = rs.sRem(ctx, RoutesKey, route.Identifier)
		return "", errors.New(fmt.Sprintf("Unable to add route details for %s, please try again", route.Identifier))
	}
	return route.Identifier, nil
}

func (rs *RedisStore) CheckRouteExistence(ctx context.Context, routeId string) (bool, error) {
	val, err := rs.sIsMember(ctx, RoutesKey, routeId)
	if err != nil {
		return false, err
	}

	if !val {
		return false, nil
	}

	//check if route details is also stored
	details, err := rs.hGetAll(ctx, fmt.Sprintf("%s%s", RouteKey, routeId))
	if err != nil {
		return false, err
	}

	if details == nil {
		return false, nil
	}
	return true, nil
}

func (rs *RedisStore) RemoveRoute(ctx context.Context, routeId string) error {
	//remove route id from set
	_, err := rs.sRem(ctx, RoutesKey, routeId)
	if err != nil {
		return err
	}

	//remove route:id hash
	hashKey := fmt.Sprintf("%s%s", RouteKey, routeId)
	keys, err := rs.client.HKeys(ctx, hashKey).Result()
	if err != nil {
		return err
	}

	err = rs.hDel(ctx, hashKey, keys)
	if err != nil {
		return err
	}
	return nil
}

func (rs *RedisStore) FetchRoutes(ctx context.Context, withDestinations bool) ([]models.Route, error) {
	routeIds, err := rs.sMembers(ctx, RoutesKey)
	if err != nil {
		return nil, err
	}

	var routeModels []models.Route
	for _, routeId := range routeIds {

		routeModel, err := rs.FetchRoute(ctx, routeId, withDestinations)
		if err != nil {
			continue
		}
		routeModels = append(routeModels, routeModel)
	}

	return routeModels, nil
}

func (rs *RedisStore) FetchRoute(ctx context.Context, routeId string, withDestinations bool) (models.Route, error) {

	routeDetails, err := rs.hGetAll(ctx, fmt.Sprintf("%s%s", RouteKey, routeId))
	if err != nil {
		return models.Route{}, err
	}

	var routeModel models.Route
	routeModel.HttpMethod, _ = enums.ParseHttpMethod(routeDetails["http_method"])
	routeModel.Name = routeDetails["name"]
	routeModel.Identifier = routeDetails["identifier"]
	routeModel.VerificationType, _ = enums.ParseVerificationType(routeDetails["verification_type"])
	routeModel.VerificationKeyLocation, _ = enums.ParseVerificationKeyLocation(routeDetails["verification_key_location"])
	routeModel.VerificationToken = routeDetails["verification_token"]
	routeModel.VerificationKeyName = routeDetails["verification_key_name"]
	routeModel.Active = routeDetails["active"] == "true"

	if withDestinations {
		destinations, err := rs.FetchDestinations(ctx, routeId)
		if err == nil {
			routeModel.Destinations = destinations
		}
	}
	return routeModel, nil
}

func (rs *RedisStore) AddDestination(ctx context.Context, routeId string, destination models.Destination) (string, error) {
	destinationsKey := fmt.Sprintf("%s:%s", DestinationsKey, routeId)
	destinationKey := fmt.Sprintf("%s%s", DestinationKey, destination.Identifier)

	//add destination ID to Set
	res, err := rs.sAdd(ctx, destinationKey, destination.Identifier)
	if err != nil {
		return "", err
	}

	if res == 0 {
		return "", errors.New(fmt.Sprintf("Destination with ID %s already exists", destination.Identifier))
	}

	//add destination details to Hash
	res, err = rs.hSet(ctx, destinationKey, destination)
	if err != nil {
		_, _ = rs.sRem(ctx, destinationsKey, destination.Identifier)
		return "", err
	}

	if res == 0 {
		_, _ = rs.sRem(ctx, destinationsKey, destination.Identifier)
		return "", errors.New(fmt.Sprintf("Unable to add destination details for %s, please try again", destination.Identifier))
	}
	return destination.Identifier, nil
}

func (rs *RedisStore) CheckDestinationExistence(ctx context.Context, routeId string, destinationId string) (bool, error) {
	val, err := rs.sIsMember(ctx, fmt.Sprintf("%s:%s", DestinationsKey, routeId), destinationId)
	if err != nil {
		return false, err
	}

	if !val {
		return false, nil
	}

	//check if destination details is also stored
	details, err := rs.hGetAll(ctx, fmt.Sprintf("%s:%s:%s", DestinationKey, routeId, destinationId))
	if err != nil {
		return false, err
	}

	if details == nil {
		return false, nil
	}
	return true, nil
}

func (rs *RedisStore) FetchDestinations(ctx context.Context, routeId string) ([]models.Destination, error) {
	destinationIds, err := rs.sMembers(ctx, fmt.Sprintf("%s:%s", DestinationsKey, routeId))
	if err != nil {
		return nil, err
	}

	var destinationModels []models.Destination
	for _, destinationId := range destinationIds {

		destinationModel, err := rs.FetchDestination(ctx, routeId, destinationId)
		if err != nil {
			continue
		}
		destinationModels = append(destinationModels, destinationModel)
	}

	return destinationModels, nil
}

func (rs *RedisStore) FetchDestination(ctx context.Context, routeId string, destinationId string) (models.Destination, error) {

	routeDetails, err := rs.hGetAll(ctx, fmt.Sprintf("%s:%s:%s", DestinationKey, routeId, destinationId))
	if err != nil {
		return models.Destination{}, err
	}

	var destinationModel models.Destination
	destinationModel.HttpMethod, _ = enums.ParseHttpMethod(routeDetails["http_method"])
	destinationModel.Identifier = routeDetails["identifier"]
	destinationModel.Active = routeDetails["active"] == "true"
	destinationModel.Url = routeDetails["url"]
	destinationModel.VerificationToken = routeDetails["verification_token"]
	return destinationModel, nil
}

func (rs *RedisStore) RemoveDestination(ctx context.Context, destinationId string) error {
	destinationsKey := fmt.Sprintf("%s:%s", DestinationsKey, destinationId)
	destinationKey := fmt.Sprintf("%s%s", DestinationKey, destinationId)

	//remove destination id from set
	_, err := rs.sRem(ctx, destinationsKey, destinationId)
	if err != nil {
		return err
	}

	//remove destination:id hash
	keys, err := rs.client.HKeys(ctx, destinationKey).Result()
	if err != nil {
		return err
	}

	err = rs.hDel(ctx, destinationKey, keys)
	if err != nil {
		return err
	}
	return nil
}

func (rs *RedisStore) get(ctx context.Context, key string) (string, error) {
	val, err := rs.client.Get(ctx, key).Result()
	if err != nil {
		return "", err
	}
	return val, nil
}

func (rs *RedisStore) set(ctx context.Context, key string, value any, ttl time.Duration) (string, error) {
	val, err := rs.client.Set(ctx, key, value, ttl).Result()
	if err != nil {
		return "", err
	}
	return val, nil
}

func (rs *RedisStore) sAdd(ctx context.Context, key string, value any) (int, error) {
	val, err := rs.client.SAdd(ctx, key, value).Result()
	if err != nil {
		return 0, err
	}
	return int(val), nil
}

func (rs *RedisStore) sRem(ctx context.Context, key string, value any) (int, error) {
	val, err := rs.client.SRem(ctx, key, value).Result()
	if err != nil {
		return 0, err
	}
	return int(val), nil
}

func (rs *RedisStore) sIsMember(ctx context.Context, key string, member interface{}) (bool, error) {
	val, err := rs.client.SIsMember(ctx, key, member).Result()
	if err != nil {
		return false, err
	}
	return val, nil
}

func (rs *RedisStore) sMembers(ctx context.Context, key string) ([]string, error) {
	val, err := rs.client.SMembers(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	return val, nil
}

func (rs *RedisStore) hSet(ctx context.Context, key string, value any) (int, error) {
	val, err := rs.client.HSet(ctx, key, value).Result()
	if err != nil {
		return 0, err
	}
	return int(val), nil
}

func (rs *RedisStore) hGetAll(ctx context.Context, key string) (map[string]string, error) {
	val, err := rs.client.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	return val, nil
}

func (rs *RedisStore) hDel(ctx context.Context, key string, fields []string) error {
	return rs.client.HDel(ctx, key, fields...).Err()
}

func NewRedisStore(client *redis.Client) *RedisStore {
	return &RedisStore{
		client: client,
	}
}
