package diagnostic

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTrace_InitTable(t *testing.T) {
	table := traceTable{}
	assert.Zero(t, len(table.tbl))
}

func TestTrace_UpdateTableX1(t *testing.T) {
	table := traceTable{}
	ctx := Context{}
	ctx.l4.hash = uint64(1234)

	table.updateTable(&ctx.l4)
	assert.Equal(t, 1, len(table.tbl))

	table.updateTable(&ctx.l4)
	assert.Equal(t, 1, len(table.tbl))
}

func TestTrace_UpdateTableX2(t *testing.T) {
	table := traceTable{}
	ctx := Context{}
	ctx.l4.hash = uint64(1234)

	table.updateTable(&ctx.l4)
	assert.Equal(t, 1, len(table.tbl))

	ctx.l4.hash = uint64(4321)
	table.updateTable(&ctx.l4)
	assert.Equal(t, 2, len(table.tbl))
}
