package decoder

import (
	"strings"

	"google.golang.org/protobuf/proto"
	"sigs.k8s.io/kpng/localsink"
	"sigs.k8s.io/kpng/pkg/api/localnetv1"
)

type Interface interface {
	// subset of localsink.Sink

	// WaitRequest see localsink.Sink#WaitRequest
	WaitRequest() (nodeName string, err error)
	// Reset see localsink.Sink#Reset
	Reset()

	// methods handling decoded values

	// SetService is called when a service is added or updated
	SetService(service *localnetv1.Service)
	// DeleteService is called when a service is deleted
	DeleteService(namespace, name string)

	// SetEndpoint is called when an endpoint is added or updated
	SetEndpoint(namespace, serviceName, key string, endpoint *localnetv1.Endpoint)
	// DeleteEndpoint is called when an endpoint is deleted
	DeleteEndpoint(namespace, serviceName, key string)
}

type Sink struct {
	Interface
}

var _ localsink.Sink = &Sink{}

func (s *Sink) Send(op *localnetv1.OpItem) (err error) {
	switch v := op.Op; v.(type) {
	case *localnetv1.OpItem_Set:
		set := op.GetSet()

		switch set.Ref.Set {
		case localnetv1.Set_ServicesSet:
			v := &localnetv1.Service{}

			err = proto.Unmarshal(set.Bytes, v)
			if err != nil {
				return
			}

			s.SetService(v)

		case localnetv1.Set_EndpointsSet:
			v := &localnetv1.Endpoint{}

			err = proto.Unmarshal(set.Bytes, v)
			if err != nil {
				return
			}

			parts := strings.Split(set.Ref.Path, "/")
			s.SetEndpoint(parts[0], parts[1], parts[2], v)

		default:
			return
		}

	case *localnetv1.OpItem_Delete:
		del := op.GetDelete()
		parts := strings.Split(del.Path, "/")

		switch del.Set {
		case localnetv1.Set_ServicesSet: // Service: namespace/name
			s.DeleteService(parts[0], parts[1])

		case localnetv1.Set_EndpointsSet: // Endpoint: namespace/name/key
			s.DeleteEndpoint(parts[0], parts[1], parts[2])

		default:
			// unknown set, ignore
		}

	case *localnetv1.OpItem_Sync:
		// ignore
		// XXX anything for the interface here?
	}

	return
}
