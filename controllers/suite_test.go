package controllers

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type ControllerSuite struct {
	suite.Suite
}

func TestSuite(t *testing.T) {
	suite.Run(t, new(ControllerSuite))
}
