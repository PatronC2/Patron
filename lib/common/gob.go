package common

import (
	"encoding/gob"

    "github.com/PatronC2/Patron/types"
)

func RegisterGobTypes() {
	for _, t := range []interface{}{
		types.Request{},
		types.ConfigurationRequest{},
		types.ConfigurationResponse{},
		types.CommandRequest{},
		types.CommandResponse{},
		types.CommandStatusRequest{},
		types.CommandStatusResponse{},
		types.KeysRequest{},
		types.KeysResponse{},
	} {
		gob.Register(t)
	}
}
