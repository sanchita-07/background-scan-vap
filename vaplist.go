package main

import (
	"k8s.io/api/admissionregistration/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

type ValidatingAdmissionPolicyListerExpansion interface{}

type ValidatingAdmissionPolicyLister interface {
	// List lists all ValidatingAdmissionPolicies in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.ValidatingAdmissionPolicy, err error)
	// Get retrieves the ValidatingAdmissionPolicy from the index for a given name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1alpha1.ValidatingAdmissionPolicy, error)
	ValidatingAdmissionPolicyListerExpansion
}

type validatingAdmissionPolicyLister struct {
	indexer cache.Indexer
}

func NewValidatingAdmissionPolicyLister(indexer cache.Indexer) ValidatingAdmissionPolicyLister {
	// validatingAdmissionPolicyLister := 
	return &validatingAdmissionPolicyLister{indexer: indexer}
}

// List lists all ValidatingAdmissionPolicies in the indexer.
func (s *validatingAdmissionPolicyLister) List(selector labels.Selector) (ret []*v1alpha1.ValidatingAdmissionPolicy, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.ValidatingAdmissionPolicy))
	})
	return ret, err
}

// Get retrieves the ValidatingAdmissionPolicy from the index for a given name.
func (s *validatingAdmissionPolicyLister) Get(name string) (*v1alpha1.ValidatingAdmissionPolicy, error) {
	obj, exists, err := s.indexer.GetByKey(name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("validatingadmissionpolicy"), name)
	}
	return obj.(*v1alpha1.ValidatingAdmissionPolicy), nil
}

// func (l *validatingAdmissionPolicyLister) List(selector labels.Selector) (ret []*v1alpha1.ValidatingAdmissionPolicy, err error) {
// 	// validatingAdmissionPolicyLister.List(selector)
// 	// ret, err = l.indexer.List(selector)
// 	// return
// 	return
// }

// func (l *validatingAdmissionPolicyLister) Get(name string) (ret *v1alpha1.ValidatingAdmissionPolicy, err error) {
// 	// validatingAdmissionPolicyLister.Get(name)
// 	// ret, err = l.indexer.ByNamespace("").Get(name)
// 	// return
// 	return
// }