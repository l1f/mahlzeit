package app

import (
	"testing"

	"github.com/l1f/mahlzeit/internal/testhelper"
	"github.com/alecthomas/assert/v2"
)

func TestApplication_GetAllRecipes(t *testing.T) {
	t.Parallel()
	app, ctx := newApp(t), testhelper.Context(t)

	for i := 0; i < 10; i++ {
		app.AddEmptyRecipe(ctx)
	}

	recipes, err := app.GetAllRecipes(ctx)
	assert.NoError(t, err)
	assert.NotZero(t, recipes)

	for _, r := range recipes {
		assert.NotZero(t, r.ID)
		assert.NotZero(t, r.Name)
	}
}

func TestApplication_GetSingleRecipe(t *testing.T) {
	t.Parallel()
	app, ctx := newApp(t), testhelper.Context(t)

	recipe := app.AddEmptyRecipe(ctx)
	step := app.AddTestStep(ctx, recipe.ID)

	ingredient, err := app.AddIngredient(ctx, t.Name())
	assert.NoError(t, err)

	err = app.AddIngredientToStep(ctx, AddIngredientToStepParams{
		StepID:       step.ID,
		IngredientID: ingredient.ID,
		Amount:       1,
		Note:         t.Name(),
	})
	assert.NoError(t, err)

	res, err := app.GetSingleRecipe(ctx, recipe.ID)
	assert.NoError(t, err)
	assert.NotZero(t, res.ID)
	assert.NotZero(t, res.Name)
	assert.NotZero(t, res.UpdatedAt)
	assert.NotZero(t, res.CreatedAt)
	assert.Equal(t, 1, len(res.Steps))
	assert.Equal(t, 1, len(res.Ingredients))
}

func TestApplication_UpdateRecipe(t *testing.T) {
	t.Parallel()
	app, ctx := newApp(t), testhelper.Context(t)

	rndmStr := testhelper.RandomString(20)

	tests := []struct {
		name     string
		updateFn func(r *Recipe)
		want     Recipe
	}{
		{
			name: "name",
			updateFn: func(r *Recipe) {
				r.Name = rndmStr
			},
			want: Recipe{
				Name: rndmStr,
			},
		},
		{
			name: "servings",
			updateFn: func(r *Recipe) {
				r.Servings = 100
			},
			want: Recipe{
				Servings:     100,
				BaseServings: 100,
			},
		},
		{
			name: "description",
			updateFn: func(r *Recipe) {
				r.Description = t.Name()
			},
			want: Recipe{
				Description: t.Name(),
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			recipe := app.AddEmptyRecipe(ctx)
			tt.updateFn(&recipe)

			err := app.UpdateRecipe(ctx, &recipe)
			assert.NoError(t, err)

			res, err := app.GetSingleRecipe(ctx, recipe.ID)
			assert.NoError(t, err)
			testhelper.PartialEqual(t, tt.want, *res)
		})
	}
}
