package controllers

import (
	"testing"

	"github.com/stretchr/testify/suite"
	extensionstsuruiov1alpha1 "github.com/tsuru/acl-operator/api/v1alpha1"
	v1alpha1 "github.com/tsuru/acl-operator/api/v1alpha1"
	"k8s.io/client-go/kubernetes/scheme"
)

type ControllerSuite struct {
	suite.Suite
}

func (suite *ControllerSuite) SetupTest() {

	var err error

	err = v1alpha1.AddToScheme(scheme.Scheme)
	suite.Require().NoError(err)

	err = extensionstsuruiov1alpha1.AddToScheme(scheme.Scheme)
	suite.Require().NoError(err)

}

func TestSuite(t *testing.T) {
	suite.Run(t, new(ControllerSuite))
}
