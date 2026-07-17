package gyre_test

import (
	"context"
	"github.com/yaop-labs/gyre"
	"testing"
)

func TestConformanceCheck(t *testing.T) {
	if err := gyre.ConformanceCheck(context.Background(), &component{}); err != nil {
		t.Fatal(err)
	}
}
