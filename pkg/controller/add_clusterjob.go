package controller

import (
	"github.com/yeochinyi/cluster-job/pkg/controller/clusterjob"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, clusterjob.Add)
}
