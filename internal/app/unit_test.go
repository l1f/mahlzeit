package app

import (
	"testing"

	"github.com/l1f/mahlzeit/internal/testhelper"
	"github.com/alecthomas/assert/v2"
)

func TestApplication_GetAllUnits(t *testing.T) {
	t.Parallel()
	app, ctx := newApp(t), testhelper.Context(t)

	for i := 0; i < 10; i++ {
		_, err := app.Queries.AddUnit(ctx, testhelper.RandomString(10))
		assert.NoError(t, err)
	}

	units, err := app.GetAllUnits(ctx)
	assert.NoError(t, err)
	assert.True(t, len(units) >= 10)
}
