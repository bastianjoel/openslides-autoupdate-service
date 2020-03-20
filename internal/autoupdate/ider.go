package autoupdate

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
)

// ErrUnknownKey ist returned from RestrictedIDs, when the requested key is not
// returned from the restricter.
var ErrUnknownKey = errors.New("key does not exist")

// RestrictedIDs implements the IDer interface by using a restricer.
type RestrictedIDs struct {
	user int
	r    Restricter
}

// ID returns the id in the key.
func (i RestrictedIDs) ID(ctx context.Context, key string) (int, error) {
	data, err := i.decodedRestricter(ctx, key)
	if err != nil {
		return 0, err
	}

	id, err := strconv.Atoi(string(data))
	if err != nil {
		return 0, fmt.Errorf("value in key %s is not an int, got: %v", key, data)
	}

	return id, nil
}

// IDList returns the ids in the key.
func (i RestrictedIDs) IDList(ctx context.Context, key string) ([]int, error) {
	data, err := i.decodedRestricter(ctx, key)
	if err != nil {
		return nil, err
	}

	var value []int
	if err := json.Unmarshal(data, &value); err != nil {
		return nil, fmt.Errorf("can not read ids from value `%s`: %w", data, err)
	}
	return value, nil
}

// GenericID returns a collection-id tuple.
func (i RestrictedIDs) GenericID(ctx context.Context, key string) (string, error) {
	data, err := i.decodedRestricter(ctx, key)
	if err != nil {
		return "", err
	}

	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return "", fmt.Errorf("can not decode generic value from restricter: %w", err)
	}
	return value, nil
}

// GenericIDs returns a list of collection-id tuple.
func (i RestrictedIDs) GenericIDs(ctx context.Context, key string) ([]string, error) {
	data, err := i.decodedRestricter(ctx, key)
	if err != nil {
		return nil, err
	}

	var values []string
	if err := json.Unmarshal(data, &values); err != nil {
		return nil, fmt.Errorf("can not decode generic-list value from restricter: %w", err)
	}
	return values, nil
}

// Template returns the strings from a template field.
func (i RestrictedIDs) Template(ctx context.Context, key string) ([]string, error) {
	return i.GenericIDs(ctx, key)
}

func (i RestrictedIDs) decodedRestricter(ctx context.Context, key string) ([]byte, error) {
	data, err := i.r.Restrict(ctx, i.user, []string{key})
	if err != nil {
		return nil, fmt.Errorf("can not restrict key %s: %w", key, err)
	}

	if _, ok := data[key]; !ok {
		return nil, ErrUnknownKey
	}
	return []byte(data[key]), nil
}
