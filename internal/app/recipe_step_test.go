package app

import (
	"net/http"
	"testing"
	"time"

	"github.com/l1f/mahlzeit/db/queries"
	"github.com/l1f/mahlzeit/internal/testhelper"
	"github.com/alecthomas/assert/v2"
	"github.com/carlmjohnson/resperr"
	"github.com/jackc/pgtype"
)

func TestApplication_AddIngredientToStep(t *testing.T) {
	t.Parallel()
	app, ctx := newApp(t), testhelper.Context(t)

	ingredient, err := app.AddIngredient(ctx, t.Name())
	assert.NoError(t, err)

	recipe := app.AddEmptyRecipe(ctx)
	step := app.AddTestStep(ctx, recipe.ID)

	tests := []struct {
		name     string
		params   AddIngredientToStepParams
		wantCode int
	}{
		{
			name: "missing step",
			params: AddIngredientToStepParams{
				StepID:       -1,
				IngredientID: ingredient.ID,
				Amount:       1,
			},
			wantCode: http.StatusNotFound,
		},
		{
			name: "missing ingredient",
			params: AddIngredientToStepParams{
				StepID:       step.ID,
				IngredientID: -1,
				Amount:       1,
			},
			wantCode: http.StatusBadRequest,
		},
		{
			name: "zero amount is ignored",
			params: AddIngredientToStepParams{
				StepID:       step.ID,
				IngredientID: ingredient.ID,
			},
			wantCode: http.StatusOK,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := app.AddIngredientToStep(ctx, tt.params)
			code := resperr.StatusCode(err)
			assert.Equal(t, tt.wantCode, code)
		})
	}
}

func TestApplication_AddStepToRecipe(t *testing.T) {
	t.Parallel()
	app, ctx := newApp(t), testhelper.Context(t)
	recipe := app.AddEmptyRecipe(ctx)

	step := app.AddTestStep(ctx, recipe.ID)
	testhelper.PartialEqual(t, Step{
		RecipeID:    recipe.ID,
		Instruction: "",
		Time:        0,
	}, step)
}

func TestApplication_DeleteIngredientFromStep(t *testing.T) {
	t.Parallel()
	app, ctx := newApp(t), testhelper.Context(t)
	recipe := app.AddEmptyRecipe(ctx)
	ingredient, _ := app.AddIngredient(ctx, t.Name())

	t.Run("missing step returns ok", func(t *testing.T) {
		err := app.DeleteIngredientFromStep(ctx, DeleteIngredientFromStepParams{
			StepID:       -1,
			IngredientID: ingredient.ID,
		})
		code := resperr.StatusCode(err)
		assert.Equal(t, http.StatusOK, code)
	})
	t.Run("missing ingredient returns ok", func(t *testing.T) {
		step := app.AddTestStep(ctx, recipe.ID)
		err := app.DeleteIngredientFromStep(ctx, DeleteIngredientFromStepParams{
			StepID:       step.ID,
			IngredientID: -1,
		})
		code := resperr.StatusCode(err)
		assert.Equal(t, http.StatusOK, code)
	})
	t.Run("ingredient is removed", func(t *testing.T) {
		step := app.AddTestStep(ctx, recipe.ID)
		err := app.AddIngredientToStep(ctx, AddIngredientToStepParams{
			StepID:       step.ID,
			IngredientID: ingredient.ID,
			Amount:       1,
		})
		assert.NoError(t, err)

		err = app.DeleteIngredientFromStep(ctx, DeleteIngredientFromStepParams{
			StepID:       step.ID,
			IngredientID: ingredient.ID,
		})
		code := resperr.StatusCode(err)
		assert.Equal(t, http.StatusOK, code)

		var count int
		err = app.DB.QueryRow(ctx, "select count(*) from step_ingredients where step_id = $1 and ingredients_id = $2", step.ID, ingredient.ID).
			Scan(&count)
		assert.NoError(t, err)
		assert.Zero(t, count)
	})
}

func TestApplication_DeleteRecipeStepByID(t *testing.T) {
	t.Parallel()
	app, ctx := newApp(t), testhelper.Context(t)
	recipe := app.AddEmptyRecipe(ctx)

	t.Run("step deletion is idempotent", func(t *testing.T) {
		err := app.DeleteRecipeStepByID(ctx, -1)
		code := resperr.StatusCode(err)
		assert.Equal(t, http.StatusOK, code)
	})
	t.Run("step is deleted", func(t *testing.T) {
		step := app.AddTestStep(ctx, recipe.ID)
		err := app.DeleteRecipeStepByID(ctx, step.ID)
		assert.NoError(t, err)

		var count int
		err = app.DB.QueryRow(ctx, "select count(*) from steps where recipe_id = $1", recipe.ID).
			Scan(&count)
		assert.NoError(t, err)
		assert.Zero(t, count)
	})
}

func TestApplication_UpdateStep(t *testing.T) {
	t.Parallel()
	app, ctx := newApp(t), testhelper.Context(t)
	recipe := app.AddEmptyRecipe(ctx)

	tests := []struct {
		name   string
		update func(s *Step)
		want   queries.Step
	}{
		{
			name: "instruction is set",
			update: func(s *Step) {
				s.Instruction = t.Name()
			},
			want: queries.Step{
				Instruction: t.Name(),
			},
		},
		{
			name: "time is set",
			update: func(s *Step) {
				s.Time = time.Minute
			},
			want: queries.Step{
				Time: pgtype.Interval{
					Microseconds: int64(time.Minute / 1000),
					Status:       pgtype.Present,
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			step := app.AddTestStep(ctx, recipe.ID)

			tt.update(&step)

			err := app.UpdateStep(ctx, step)
			assert.NoError(t, err)

			dbStep, err := app.Queries.GetStepByID(ctx, int64(step.ID))
			assert.NoError(t, err)
			testhelper.PartialEqual(t, tt.want, dbStep)
		})
	}
}

func TestApplication_GetStepByID(t *testing.T) {
	app, ctx := newApp(t), testhelper.Context(t)

	t.Run("not existing step returns not found error", func(t *testing.T) {
		t.Parallel()
		_, err := app.GetStepByID(ctx, -1)
		assert.Error(t, err)
		assert.Equal(t, http.StatusNotFound, resperr.StatusCode(err))
	})
	t.Run("step is returned", func(t *testing.T) {
		t.Parallel()
		recipe := app.AddEmptyRecipe(ctx)
		testStep := app.AddTestStep(ctx, recipe.ID)

		got, err := app.GetStepByID(ctx, testStep.ID)
		assert.NoError(t, err)
		assert.Equal(t, testStep, got)
	})
}
