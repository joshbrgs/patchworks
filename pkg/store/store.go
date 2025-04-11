package store

import (
	"context"
	"encoding/json"
	"fmt"

	patchesv1alpha1 "github.com/bigideaslearning/patchworks/api/v1alpha1"
	"github.com/bigideaslearning/patchworks/pkg/utils"
	"github.com/redis/go-redis/v9"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/uuid"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type Store interface {
	Get(key string) (*patchesv1alpha1.TargetRef, error)
	SetOriginal(target patchesv1alpha1.TargetRef) (key string, err error)
}

type RedisStore struct {
	ctx context.Context
	db  *redis.Client
	c   client.Client
}

func NewStore(ctx context.Context, c client.Client, db *redis.Client) *RedisStore {
	return &RedisStore{db: db, c: c, ctx: ctx}
}

func (s *RedisStore) Get(key string) (*unstructured.Unstructured, error) {
	log := log.FromContext(s.ctx)

	res, err := s.db.Get(s.ctx, key).Result()

	if err == redis.Nil {
		return nil, fmt.Errorf("Key was not found: %s", key)
	} else if err != nil {
		return nil, err
	}

	log.Info("Retrieved Patch from store", "Redis", res)

	var obj unstructured.Unstructured
	if err := json.Unmarshal([]byte(res), &obj); err != nil {
		return nil, fmt.Errorf("failed to unmarshal resource: %w", err)
	}

	return &obj, nil

}

func (s *RedisStore) SetOriginal(target patchesv1alpha1.TargetRef) (key string, err error) {
	key = string(uuid.NewUUID())
	log := log.FromContext(s.ctx)

	targetObj, err := utils.GetResource(s.ctx, s.c, target)
	if err != nil {
		return "", fmt.Errorf("failed to get resource: %w", err)
	}

	jsonData, err := json.Marshal(targetObj)
	if err != nil {
		log.Error(err, "Error marshalling JSON")
		return "", fmt.Errorf("failed to store resource: %w", err)
	}

	if err := s.db.Set(s.ctx, key, jsonData, 0).Err(); err != nil {
		return "", err
	}

	return key, nil
}
