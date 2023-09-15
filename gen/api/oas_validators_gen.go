// Code generated by ogen, DO NOT EDIT.

package api

import (
	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen/validate"
)

func (s Clusters) Validate() error {
	alias := ([]Cluster)(s)
	if alias == nil {
		return errors.New("nil is invalid value")
	}
	return nil
}

func (s *ListClusterResponse) Validate() error {
	var failures []validate.FieldError
	if err := func() error {
		if err := s.Clusters.Validate(); err != nil {
			return err
		}
		return nil
	}(); err != nil {
		failures = append(failures, validate.FieldError{
			Name:  "clusters",
			Error: err,
		})
	}
	if len(failures) > 0 {
		return &validate.Error{Fields: failures}
	}
	return nil
}
