package tests

import (
	"github.com/goravel/framework/testing"

	"missfit/bootstrap"
)

func init() {
	bootstrap.Boot()
}

type TestCase struct {
	testing.TestCase
}
