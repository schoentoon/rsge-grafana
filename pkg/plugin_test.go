package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gitlab.com/schoentoon/rs-tools/lib/ge"
)

func TestGraphToTTL(t *testing.T) {
	now := time.Date(2021, time.January, 5, 18, 36, 0, 0, time.UTC)

	graph := &ge.Graph{
		Graph: make(map[time.Time]int),
	}

	graph.Graph[time.Date(2021, time.January, 5, 0, 0, 0, 0, time.UTC)] = 0

	assert.Equal(t, time.Duration(time.Hour*5+time.Minute*24), graphToTTL(now, graph))
}
