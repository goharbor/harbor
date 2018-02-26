package source

import (
	"github.com/vmware/harbor/src/replication"
	"github.com/vmware/harbor/src/replication/registry"
)

//Sourcer is used to manage and/or handle all the artifacts and information related with source registry.
//All the things with replication source should be covered in this object.
type Sourcer struct {
	//Keep the adaptors we support now
	adaptors map[string]registry.Adaptor
}

//NewSourcer is the constructor of Sourcer
func NewSourcer() *Sourcer {
	return &Sourcer{
		adaptors: make(map[string]registry.Adaptor),
	}
}

//Init will do some initialization work like registrying all the adaptors we support
func (sc *Sourcer) Init() {
	//Register Harbor adaptor
	sc.adaptors[replication.AdaptorKindHarbor] = &registry.HarborAdaptor{}
}

//GetAdaptor returns the required adaptor with the specified kind.
//If no adaptor with the specified kind existing, nil will be returned.
func (sc *Sourcer) GetAdaptor(kind string) registry.Adaptor {
	if len(kind) == 0 {
		return nil
	}

	return sc.adaptors[kind]
}
