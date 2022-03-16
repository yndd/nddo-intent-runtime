package intent

import (
	"context"
	"fmt"

	"github.com/yndd/nddo-runtime/pkg/resource"
)

type Methods interface {
	Deploy(ctx context.Context, mg resource.Managed, labels map[string]string) error
	Destroy(ctx context.Context, mg resource.Managed, labels map[string]string) error
	List(ctx context.Context, mg resource.Managed, resources map[string]map[string]struct{}) (map[string]map[string]struct{}, error)
	Validate(ctx context.Context, mg resource.Managed, resources map[string]map[string]struct{}) (map[string]map[string]struct{}, error)
	Delete(ctx context.Context, mg resource.Managed, resources map[string]map[string]struct{}) error
}

type Instances interface {
	// methods children
	NewInstance(c resource.ClientApplicator, name string, initFn InstanceInitFunc) Instance
	GetInstances() map[string]Instance
	Print(string) error
}

type Data interface {
	Get() interface{}
	Print() error
}

type Instance interface {
	// instance data methods
	Data
	// intent methods
	Methods
}

type Intent interface {
	// method instances children
	Instances
	// intent methods
	Methods
}

func New(c resource.ClientApplicator) Intent {
	return &intent{
		// k8s client
		client: c,
		// parent is nil/root
		// children
		instance: make(map[string]Instance),
		// data key
	}
}

type intent struct {
	// k8s client
	client resource.ClientApplicator
	// parent is nil/root
	// children
	instance map[string]Instance
	// data is nil
}

func NewInstance(c resource.ClientApplicator, p Intent, name string) Instance {
	return nil
}

type InstanceInitFunc func(c resource.ClientApplicator, p Intent, name string) Instance


func (x *intent) NewInstance(c resource.ClientApplicator, name string, initFn InstanceInitFunc) Instance {
	if _, ok := x.instance[name]; !ok {
		x.instance[name] = initFn(c, x, name)
	}
	return x.instance[name]
}

func (x *intent) GetInstances() map[string]Instance {
	return x.instance
}

func (x *intent) Print(crName string) error {
	fmt.Printf("intent information: %s\n", crName)
	for _, d := range x.GetInstances() {
		if err := d.Print(); err != nil {
			return err
		}
	}
	return nil
}

func (x *intent) Deploy(ctx context.Context, mg resource.Managed, labels map[string]string) error {
	for _, i := range x.GetInstances() {
		if err := i.Deploy(ctx, mg, labels); err != nil {
			return err
		}
	}
	return nil
}

func (x *intent) Destroy(ctx context.Context, mg resource.Managed, labels map[string]string) error {
	for _, i := range x.GetInstances() {
		if err := i.Destroy(ctx, mg, labels); err != nil {
			return err
		}
	}
	return nil
}

func (x *intent) List(ctx context.Context, mg resource.Managed, resources map[string]map[string]struct{}) (map[string]map[string]struct{}, error) {
	var err error
	for _, i := range x.GetInstances() {
		resources, err = i.List(ctx, mg, resources)
		if err != nil {
			return nil, err
		}
	}
	return resources, nil
}

func (x *intent) Validate(ctx context.Context, mg resource.Managed, resources map[string]map[string]struct{}) (map[string]map[string]struct{}, error) {
	var err error
	for _, i := range x.GetInstances() {
		resources, err = i.Validate(ctx, mg, resources)
		if err != nil {
			return nil, err
		}
	}
	return resources, nil
}

func (x *intent) Delete(ctx context.Context, mg resource.Managed, resources map[string]map[string]struct{}) error {
	for _, i := range x.GetInstances() {
		if err := i.Delete(ctx, mg, resources); err != nil {
			return err
		}
	}
	return nil
}
