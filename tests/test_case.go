package tests

import (
	"github.com/goravel/framework/testing"

	"lumos/bootstrap"
)

func init() {
	bootstrap.Boot()
}

type TestCase struct {
	testing.TestCase
}
