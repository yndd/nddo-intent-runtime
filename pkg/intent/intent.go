package intent

import (
	"context"
	"sync"

	"github.com/yndd/nddo-runtime/pkg/resource"
)

type Intent interface {
	Deploy(ctx context.Context, mg resource.Managed, labels map[string]string) error
	Destroy(ctx context.Context, mg resource.Managed, labels map[string]string) error
	List(ctx context.Context, mg resource.Managed, resources map[string]map[string]struct{}) (map[string]map[string]struct{}, error)
	Validate(ctx context.Context, mg resource.Managed, resources map[string]map[string]struct{}) (map[string]map[string]struct{}, error)
	Delete(ctx context.Context, mg resource.Managed, resources map[string]map[string]struct{}) error
	GetData() interface{}
}

func New(c resource.ClientApplicator, name string) *Compositeintent {
	return &Compositeintent{
		name: name,
		// k8s client
		client: c,
		// parent is nil/root
		// children
		intents: make(map[string]Intent),
		// data key
	}
}

type Compositeintent struct {
	name string
	// k8s client
	client resource.ClientApplicator
	// parent is nil/root
	// children
	m       sync.Mutex
	intents map[string]Intent
	// data is nil
}

func (x *Compositeintent) AddChild(name string, i Intent) {
	x.m.Lock()
	defer x.m.Unlock()
	if _, ok := x.intents[name]; !ok {
		x.intents[name] = i
	}
}

func (x *Compositeintent) GetChildData(name string) interface{} {
	x.m.Lock()
	defer x.m.Unlock()
	return x.intents[name].GetData()
}

func (x *Compositeintent) GetData() interface{} {
	x.m.Lock()
	defer x.m.Unlock()
	d := make(map[string]interface{})
	for name, i := range x.intents {
		d[name] = i.GetData()
	}
	return d
}

func (x *Compositeintent) Deploy(ctx context.Context, mg resource.Managed, labels map[string]string) error {
	x.m.Lock()
	defer x.m.Unlock()
	for _, i := range x.intents {
		if err := i.Deploy(ctx, mg, labels); err != nil {
			return err
		}
	}
	return nil
}

func (x *Compositeintent) Destroy(ctx context.Context, mg resource.Managed, labels map[string]string) error {
	x.m.Lock()
	defer x.m.Unlock()
	for _, i := range x.intents {
		if err := i.Destroy(ctx, mg, labels); err != nil {
			return err
		}
	}
	return nil
}

func (x *Compositeintent) List(ctx context.Context, mg resource.Managed, resources map[string]map[string]struct{}) (map[string]map[string]struct{}, error) {
	x.m.Lock()
	defer x.m.Unlock()
	var err error
	for _, i := range x.intents {
		resources, err = i.List(ctx, mg, resources)
		if err != nil {
			return nil, err
		}
	}
	return resources, nil
}

func (x *Compositeintent) Validate(ctx context.Context, mg resource.Managed, resources map[string]map[string]struct{}) (map[string]map[string]struct{}, error) {
	x.m.Lock()
	defer x.m.Unlock()
	var err error
	for _, i := range x.intents {
		resources, err = i.Validate(ctx, mg, resources)
		if err != nil {
			return nil, err
		}
	}
	return resources, nil
}

func (x *Compositeintent) Delete(ctx context.Context, mg resource.Managed, resources map[string]map[string]struct{}) error {
	x.m.Lock()
	defer x.m.Unlock()
	for _, i := range x.intents {
		if err := i.Delete(ctx, mg, resources); err != nil {
			return err
		}
	}
	return nil
}
