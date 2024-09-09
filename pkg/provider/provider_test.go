package provider

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	bolt "go.etcd.io/bbolt"
)

func TestNewProvider(t *testing.T) {
	p := NewProvider(nil)
	if reflect.TypeOf(p) != reflect.TypeOf(&WhitelistProvider{}) {
		t.Fatalf("expected NewProvider to return a &from_env.FromEnvProvider, got %T", p)
	}
}

func TestWithPayload(t *testing.T) {
	db, err := bolt.Open("config.test.db", 0600, nil)
	if err != nil {
		t.Fatalf("unable to open test db %v", err)
	}
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("whitelist"))
		if err != nil {
			t.Fatalf("unable to create test bucket %v", err)
		}
		for _, v := range []string{"foo", "bar", "baz"} {
			key := fmt.Sprintf("name:%s", v)
			if err := b.Put([]byte(key), []byte(v)); err != nil {
				t.Fatalf("unable to insert to bucket %v", err)
				continue
			}
		}
		return nil
	})

	p := NewProvider(db)

	evalCtx := map[string]interface{}{
		"key":    "name:",
		"msisdn": "foo",
	}

	res := p.BooleanEvaluation(context.Background(), "whitelist", false, evalCtx)

	if res.Error() != nil {
		t.Fatalf("expected no error, got %s", res.Error())
	}

	if !res.Value {
		t.Fatalf("expected true value, got false")
	}

	if res.Variant != "whitelist" {
		t.Fatalf("expected foo variant, got %s", res.Variant)
	}
}
