/*
Module created for easier unit testing

*/
package listserialports

import (
	"path/filepath"
)

type SymlinkEvaluator interface {
	Eval(string) (string, error)
}

type FilepathLinkEvaluator struct{}

func (p FilepathLinkEvaluator) Eval(filename string) (string, error) {
	result, err := filepath.EvalSymlinks(filename)
	return result, err
}
