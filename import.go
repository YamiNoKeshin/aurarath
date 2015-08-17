package aurarath

import (
	"github.com/joernweissenborn/aurarath/config"
	"github.com/joernweissenborn/aurarath/network/node"
)

type Import struct {
	AppKey *AppKey
	node  *node.Node
}

func NewImport(config *config.Config, key *AppKey) *Import {
	imp := Import{
		AppKey: key,
	}

	return &imp
}

func (*Import) AddFunction(fkt *Function)

func (*Import) RemoveFunction(name string)

func (*Import) UpdateTags(tags []string)

func (*Import) Call(req *Request) (res *Result)

func (*Import) CallAll(req *Request) (res *Result)

func (*Import) Trigger(req *Request)

func (*Import) TriggerAll(req *Request)

func (*Import) Remove()
