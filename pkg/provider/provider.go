package provider

import (
	"context"
	"fmt"

	"github.com/open-feature/go-sdk/openfeature"

	bolt "go.etcd.io/bbolt"
)

type WhitelistProvider struct {
	db *bolt.DB
}

func NewProvider(db *bolt.DB) *WhitelistProvider {
	p := &WhitelistProvider{
		db,
	}
	return p
}

// Required: Methods below implements openfeature.FeatureProvider interface
// This is the core interface implementation required from a provider
// Metadata returns the metadata of the provider
func (i *WhitelistProvider) Metadata() openfeature.Metadata {
	return openfeature.Metadata{
		Name: "WhitelistProvider",
	}
}

// Hooks returns a collection of openfeature.Hook defined by this provider
func (i *WhitelistProvider) Hooks() []openfeature.Hook {
	// Hooks that should be included with the provider
	return []openfeature.Hook{}
}

// BooleanEvaluation returns a boolean flag
func (i *WhitelistProvider) BooleanEvaluation(ctx context.Context,
	flag string, defaultValue bool,
	evalCtx openfeature.FlattenedContext,
) openfeature.BoolResolutionDetail {

	res := i.resolveFlag(flag, defaultValue, evalCtx)
	v, ok := res.Value.(bool)

	if !ok {
		return openfeature.BoolResolutionDetail{
			Value: defaultValue,
			ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
				ResolutionError: openfeature.NewTypeMismatchResolutionError(""),
				Reason:          openfeature.ErrorReason,
			},
		}
	}

	return openfeature.BoolResolutionDetail{
		Value:                    v,
		ProviderResolutionDetail: res.ProviderResolutionDetail,
	}
}

// StringEvaluation returns a string flag
func (i *WhitelistProvider) StringEvaluation(ctx context.Context, flag string, defaultValue string, evalCtx openfeature.FlattenedContext) openfeature.StringResolutionDetail {
	return openfeature.StringResolutionDetail{}
}

// FloatEvaluation returns a float flag
func (i *WhitelistProvider) FloatEvaluation(ctx context.Context, flag string, defaultValue float64, evalCtx openfeature.FlattenedContext) openfeature.FloatResolutionDetail {
	return openfeature.FloatResolutionDetail{}
}

// IntEvaluation returns an int flag
func (i *WhitelistProvider) IntEvaluation(ctx context.Context, flag string, defaultValue int64, evalCtx openfeature.FlattenedContext) openfeature.IntResolutionDetail {
	return openfeature.IntResolutionDetail{}
}

// ObjectEvaluation returns an object flag
func (i *WhitelistProvider) ObjectEvaluation(ctx context.Context, flag string, defaultValue interface{}, evalCtx openfeature.FlattenedContext) openfeature.InterfaceResolutionDetail {
	return openfeature.InterfaceResolutionDetail{}
}

// Optional: openfeature.StateHandler implementation
// Providers can opt-in for initialization & shutdown behavior by implementing this interface

// Init holds initialization logic of the provider
func (i *WhitelistProvider) Init(evaluationContext openfeature.EvaluationContext) error {
	return nil
}

// Status expose the status of the provider
func (i *WhitelistProvider) Status() openfeature.State {
	// The state is typically set during initialization.
	return openfeature.ReadyState
}

// Shutdown define the shutdown operation of the provider
func (i *WhitelistProvider) Shutdown() {
	// code to shutdown your provider
}

func (i *WhitelistProvider) resolveFlag(flagKey string,
	defaultValue interface{},
	evalCtx openfeature.FlattenedContext,
) openfeature.InterfaceResolutionDetail {
	var value bool

	e := i.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(flagKey))
		key := evalCtx["key"]
		if key == nil {
			return fmt.Errorf("key not found")
		}
		msisdn := evalCtx["msisdn"]
		if msisdn == nil {
			return fmt.Errorf("msisdn not found")
		}
		v := b.Get([]byte(fmt.Sprintf("%s%s", key.(string), msisdn.(string))))
		if v != nil {
			value = true
		} else {
			value = false
		}
		return nil
	})

	if e != nil {
		e := openfeature.NewGeneralResolutionError(e.Error())
		return openfeature.InterfaceResolutionDetail{
			Value: defaultValue,
			ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
				Variant:         "default",
				ResolutionError: e,
				Reason:          openfeature.ErrorReason,
			},
		}
	}

	return openfeature.InterfaceResolutionDetail{
		Value: value,
		ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
			Variant: "whitelist",
		},
	}
}
