package app

import (
	"net/http"
	"testing"

	"github.com/l1f/mahlzeit/internal/testhelper"
	"github.com/alecthomas/assert/v2"
	"github.com/carlmjohnson/resperr"
)

func TestApplication_GetAllIngredients(t *testing.T) {
	t.Parallel()
	ctx := testhelper.Context(t)
	app := newApp(t)

	newID, err := app.Queries.AddIngredient(ctx, t.Name())
	assert.NoError(t, err, "adding new ingredient")

	res, err := app.GetAllIngredients(ctx)
	assert.NoError(t, err, "fetching all ingredients")

	hasIngredient := false
	for _, i := range res {
		if i.ID == int(newID) {
			hasIngredient = true
			break
		}
	}

	assert.True(t, hasIngredient, "should have ingredient %d in list", newID)
}

func TestApplication_GetIngredient(t *testing.T) {
	t.Parallel()
	app, ctx := newApp(t), testhelper.Context(t)

	t.Run("existing ingredient is returned", func(t *testing.T) {
		newIngredient, err := app.AddIngredient(ctx, testhelper.RandomString(10))
		assert.NoError(t, err)

		ingredient, err := app.GetIngredient(ctx, newIngredient.ID)
		assert.NoError(t, err)
		assert.Equal(t, ingredient, newIngredient)
	})
	t.Run("missing ingredient returns not found error", func(t *testing.T) {
		_, err := app.GetIngredient(ctx, -1)
		assert.Equal(t, resperr.StatusCode(err), http.StatusNotFound)
	})
}

func TestApplication_AddIngredient(t *testing.T) {
	t.Parallel()
	app, ctx := newApp(t), testhelper.Context(t)

	t.Run("new ingredient is added", func(t *testing.T) {
		_, err := app.AddIngredient(ctx, testhelper.RandomString(10))
		assert.NoError(t, err)
	})
	t.Run("existing ingredient is returned", func(t *testing.T) {
		name := testhelper.RandomString(10)
		first, err := app.AddIngredient(ctx, name)
		assert.NoError(t, err, "no error on first time")

		second, err := app.AddIngredient(ctx, name)
		assert.NoError(t, err, "no error on second time")
		assert.Equal(t, first, second)
	})
}
