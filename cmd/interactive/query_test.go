package interactive

import (
	"testing"

	"github.com/aerospike/avs-client-go"
	"github.com/stretchr/testify/assert"
)

func TestQuerySelection(t *testing.T) {
	// Create test data
	neighbors := createTestNeighbors()

	// Create the model
	model := NewQueryModel(neighbors, 5, 50, nil)

	// Verify initial state
	assert.Equal(t, -1, model.selected)
	assert.False(t, model.queryRequested)

	// Directly set the selected record
	model.selected = 0
	assert.Equal(t, 0, model.selected)

	// Directly set the query requested flag
	model.queryRequested = true
	assert.True(t, model.queryRequested)

	// Verify that GetSelectedVectorForQuery returns the correct neighbor
	selectedNeighbor, ok := model.GetSelectedVectorForQuery()
	assert.True(t, ok)
	assert.NotNil(t, selectedNeighbor)
	assert.Equal(t, "test-key-1", selectedNeighbor.Key)
}

func TestMultipleQueryChain(t *testing.T) {
	// This test simulates the flow from the cmd/query.go file
	// that supports multiple requery operations

	// Create initial test data
	originalNeighbors := createTestNeighbors()

	// First query cycle
	model1 := NewQueryModel(originalNeighbors, 5, 50, nil)
	model1.selected = 0
	model1.queryRequested = true

	// Get the selected neighbor for the first requery
	selectedNeighbor1, ok := model1.GetSelectedVectorForQuery()
	assert.True(t, ok)
	assert.Equal(t, "test-key-1", selectedNeighbor1.Key)

	// Simulate a second query with new results (normally this would be
	// the result of a new vector search using the previously selected vector)
	secondQueryNeighbors := []*avs.Neighbor{
		originalNeighbors[1], // Just reuse the second original neighbor for simplicity
	}

	// Second query cycle
	model2 := NewQueryModel(secondQueryNeighbors, 5, 50, nil)
	model2.selected = 0
	model2.queryRequested = true

	// Get the selected neighbor for the second requery
	selectedNeighbor2, ok := model2.GetSelectedVectorForQuery()
	assert.True(t, ok)
	assert.Equal(t, "test-key-2", selectedNeighbor2.Key)

	// Test the ability to exit the requery chain
	model3 := NewQueryModel(originalNeighbors, 5, 50, nil)
	// Don't set queryRequested flag this time
	model3.selected = 0

	// This should return false because queryRequested is not set
	_, ok = model3.GetSelectedVectorForQuery()
	assert.False(t, ok)
}

// Helper function to create test neighbors
func createTestNeighbors() []*avs.Neighbor {
	set := "test-set"
	return []*avs.Neighbor{
		{
			Key:       "test-key-1",
			Distance:  0.1,
			Namespace: "test-namespace",
			Set:       &set,
			Record: &avs.Record{
				Generation: 1,
				Data: map[string]interface{}{
					"vector": []float32{0.1, 0.2, 0.3},
					"text":   "This is a test document",
				},
			},
		},
		{
			Key:       "test-key-2",
			Distance:  0.2,
			Namespace: "test-namespace",
			Set:       &set,
			Record: &avs.Record{
				Generation: 1,
				Data: map[string]interface{}{
					"vector": []float32{0.4, 0.5, 0.6},
					"text":   "This is another test document",
				},
			},
		},
	}
}
