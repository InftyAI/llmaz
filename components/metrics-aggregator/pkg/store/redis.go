/*
Copyright 2025 The InftyAI Team.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package store

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	MAX_ELEMENTS    = 5
	EXPIRATION_TIME = 5 * time.Minute
)

type RedisStore struct {
	// ctx is the context for the Redis client
	ctx    context.Context
	client *redis.Client
}

func NewRedisStore(ctx context.Context, host string) *RedisStore {
	client := redis.NewClient(&redis.Options{
		Addr: host + ":6379",
	})

	return &RedisStore{
		ctx:    ctx,
		client: client,
	}
}

// The data will be stored in two ways:
// - Sorted set with the LeastLatency:<modelName> as the key and namespacedName as the member.
// - Common set with the namespacedName as the key and current time as the value.
// It generally looks like:
//
//	ZADD LeastLatency::<ModelName> 0.5 default/fake-pod
//	SET default/fake-pod "2025-05-12T06:16:27Z"
func (r *RedisStore) Insert(ctx context.Context, key string, score float64, member string) error {
	// No need to store if the score is 0.
	if score == 0 {
		return nil
	}

	err := r.client.ZAdd(ctx, key, redis.Z{
		Score:  score,
		Member: member,
	}).Err()
	if err != nil {
		return err
	}

	// We only keep part of the elements.
	err = r.client.ZRemRangeByRank(ctx, key, MAX_ELEMENTS, -1).Err()
	if err != nil {
		return err
	}

	currentTime := time.Now().UTC().Format(time.RFC3339)
	// Set to expire in 5 minutes.
	err = r.client.Set(ctx, member, currentTime, EXPIRATION_TIME).Err()
	if err != nil {
		return err
	}
	return nil
}

// Remove the member from the sorted set and delete the key.
func (r *RedisStore) Remove(ctx context.Context, key string, member string) error {
	if err := r.client.Del(ctx, member).Err(); err != nil {
		return err
	}
	if err := r.client.ZRem(ctx, key, member).Err(); err != nil {
		return err
	}
	return nil
}
