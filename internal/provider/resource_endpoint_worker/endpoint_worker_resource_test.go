package resource_endpoint_worker

import (
	"testing"
)

func TestEndpointWorkerResourceExists(t *testing.T) {
	// Basic test to verify the resource can be instantiated
	r := NewEndpointWorkerResource()
	if r == nil {
		t.Fatal("Failed to create endpoint worker resource")
	}
}
