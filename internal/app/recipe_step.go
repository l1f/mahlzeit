package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/l1f/mahlzeit/db/queries"
	"github.com/l1f/mahlzeit/internal/pghelper"
	"github.com/carlmjohnson/resperr"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
)

// AddStepToRecipe adds a step to the recipe.
func (app *Application) AddStepToRecipe(ctx context.Context, recipeID int, s *Step) error {
	id, err := app.Queries.AddNewStep(ctx, queries.AddNewStepParams{
		RecipeID:    int64(recipeID),
		Instruction: s.Instruction,
		Time:        pghelper.Interval(s.Time),
	})
	if err != nil {
		return fmt.Errorf("adding step in database for recipe %d: %w", id, err)
	}

	s.RecipeID = recipeID
	s.ID = int(id)

	return nil
}

// GetStepByID returns a step by its ID. If it does not exist, a not found error is returned.
func (app *Application) GetStepByID(ctx context.Context, id int) (Step, error) {
	step, err := app.Queries.GetStepByID(ctx, int64(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Step{}, resperr.WithStatusCode(err, http.StatusNotFound)
		}
		return Step{}, fmt.Errorf("querying step %d: %w", id, err)
	}

	return Step{
		ID:          int(step.ID),
		RecipeID:    int(step.RecipeID),
		Instruction: step.Instruction,
		Time:        pghelper.ToDuration(step.Time),
	}, nil
}

// UpdateStep updates an existing step.
func (app *Application) UpdateStep(ctx context.Context, s Step) error {
	if err := app.Queries.UpdateStepByID(ctx, queries.UpdateStepByIDParams{
		ID:          int64(s.ID),
		Instruction: s.Instruction,
		Time:        pghelper.Interval(s.Time),
	}); err != nil {
		return fmt.Errorf("updating step %d: %w", s.ID, err)
	}

	return nil
}

// DeleteRecipeStepByID deletes a recipe step by its ID.
// This is an idempotent action, if the step is already deleted, no error is returned.
func (app *Application) DeleteRecipeStepByID(ctx context.Context, id int) error {
	if err := app.Queries.DeleteStepByID(ctx, int64(id)); err != nil {
		return fmt.Errorf("deleting step %d: %w", id, err)
	}
	return nil
}

type AddIngredientToStepParams struct {
	StepID       int
	IngredientID int
	UnitID       *int
	Amount       float64
	Note         string
}

// AddIngredientToStep adds an ingredient to a step.
func (app *Application) AddIngredientToStep(ctx context.Context, params AddIngredientToStepParams) error {
	if err := app.Queries.AddIngredientToStep(ctx, queries.AddIngredientToStepParams{
		StepID:        int64(params.StepID),
		IngredientsID: int64(params.IngredientID),
		UnitID:        int64(valueOrDefault(params.UnitID)),
		Amount:        pghelper.Numeric(params.Amount),
		Note:          params.Note,
	}); err != nil {
		var pgerr *pgconn.PgError
		if !errors.As(err, &pgerr) {
			return fmt.Errorf("adding ingredient to step %d: %w", params.StepID, err)
		}

		switch pgerr.ConstraintName {
		case "step_ingredients_step_id_fkey":
			return resperr.WithCodeAndMessage(err, http.StatusNotFound, "step does not exist")
		case "step_ingredients_ingredients_id_fkey":
			return resperr.WithCodeAndMessage(err, http.StatusBadRequest, "ingredient does not exist")
		}

		return fmt.Errorf("adding ingredient to step %d: %w", params.StepID, err)
	}

	return nil
}

type DeleteIngredientFromStepParams struct {
	StepID       int
	IngredientID int
}

// DeleteIngredientFromStep deletes an ingredient from a step. The ingredient itself is unaffected.
// This action is idempotent, if the step is already deleted, no error is returned.
func (app *Application) DeleteIngredientFromStep(ctx context.Context, params DeleteIngredientFromStepParams) error {
	if err := app.Queries.DeleteIngredientFromStep(ctx, queries.DeleteIngredientFromStepParams{
		StepID:        int64(params.StepID),
		IngredientsID: int64(params.IngredientID),
	}); err != nil {
		return fmt.Errorf("deleting ingredient %d from step %d: %w", params.IngredientID, params.StepID, err)
	}

	return nil
}

// valueOrDefault returns the value for v or the default value of T otherwise.
func valueOrDefault[T comparable](v *T) T {
	var res T
	if v != nil {
		res = *v
	}
	return res
}
