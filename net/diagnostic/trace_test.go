package diagnostic

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTrace_InitTable(t *testing.T) {
	table := TraceTable{}
	assert.Zero(t, len(table.tbl))
}

func TestTrace_UpdateTableX1(t *testing.T) {
	table := TraceTable{}
	ctx := Context{}
	ctx.srcSum = uint64(1234)
	ctx.dstSum = uint64(4321)

	table.updateTable(&ctx)
	assert.Equal(t, 1, len(table.tbl))

	table.updateTable(&ctx)
	assert.Equal(t, 1, len(table.tbl))
}

func TestTrace_UpdateTableX2(t *testing.T) {
	table := TraceTable{}
	ctx := Context{}
	ctx.srcSum = uint64(1234)
	ctx.dstSum = uint64(4321)

	table.updateTable(&ctx)
	assert.Equal(t, 1, len(table.tbl))

	ctx.srcSum = uint64(5678)
	ctx.dstSum = uint64(4321)
	table.updateTable(&ctx)
	assert.Equal(t, 2, len(table.tbl))
}
