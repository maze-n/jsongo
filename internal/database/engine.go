package database

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
)

type Store struct {
	mu   sync.RWMutex
	data map[string][]map[string]any
}

func NewStore(r io.Reader) (*Store, error) {
	var raw map[string]json.RawMessage

	decoder := json.NewDecoder(r)

	if err := decoder.Decode(&raw); err != nil {
		return nil, fmt.Errorf("decode database: %w", err)
	}

	// Reject trailing JSON values or malformed trailing content.
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return nil, fmt.Errorf("decode database: unexpected additional JSON value")
		}
		return nil, fmt.Errorf("decode database: trailing content: %w", err)
	}

	if raw == nil {
		return nil, fmt.Errorf("database root must be a JSON object")
	}

	data := make(map[string][]map[string]any, len(raw))

	for collection, value := range raw {
		var records []map[string]any

		if err := json.Unmarshal(value, &records); err != nil {
			return nil, fmt.Errorf(
				"invalid collection %q: expected array of objects: %w",
				collection,
				err,
			)
		}

		if records == nil {
			return nil, fmt.Errorf(
				"invalid collection %q: expected an array, got null",
				collection,
			)
		}

		for i, record := range records {
			if record == nil {
				return nil, fmt.Errorf(
					"invalid collection %q: record at index %d must be an object",
					collection,
					i,
				)
			}
		}

		data[collection] = records
	}

	return &Store{
		data: data,
	}, nil
}

func LoadStore(filename string) (*Store, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("open database file %q: %w", filename, err)
	}
	defer file.Close()

	store, err := NewStore(file)
	if err != nil {
		return nil, fmt.Errorf("load database file %q: %w", filename, err)
	}

	return store, err
}
