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

import "context"

// Store represents the interface for the backend store, like Redis.
// In metrics-aggregator, we use Store to save the metrics and
// In AI gateway, we use Store to fetch the metrics for smart routing.
// They're paired with each other. Each time you want to add a new store,
// you need to implement in both sides.
//
// Note:
// The store interface functions is derived from Redis,
// change it once we have more stores.
type Store interface {
	Insert(ctx context.Context, key string, score float64, member string) error
	Remove(ctx context.Context, key string, member string) error
}
