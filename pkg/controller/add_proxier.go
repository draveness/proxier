package controller

import (
	"github.com/draveness/proxier/pkg/controller/proxier"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, proxier.Add)
}
