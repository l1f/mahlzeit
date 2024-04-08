package app

import (
	"context"
	"fmt"
	"time"

	"github.com/l1f/mahlzeit/db/queries"
	"github.com/jackc/pgtype"
)

func (app *Application) GetAllRecipes(ctx context.Context) ([]ListEntry, error) {
	dbResult, err := app.Queries.GetAllRecipesByName(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetching recipes from database: %w", err)
	}

	var res []ListEntry
	for _, row := range dbResult {
		res = append(res, ListEntry{
			ID:   int(row.ID),
			Name: row.Name,
		})
	}

	return res, nil
}

// GetSingleRecipe returns a recipe by its ID.
func (app *Application) GetSingleRecipe(ctx context.Context, id int) (*Recipe, error) {
	// TODO: execute the following queries in a transaction
	base, err := app.Queries.GetRecipeByID(ctx, int64(id))
	if err != nil {
		return nil, fmt.Errorf("querying recipe %d: %w", id, err)
	}

	res := &Recipe{
		ID:                  int(base.ID),
		Name:                base.Name,
		Description:         base.Description,
		CreatedAt:           base.CreatedAt,
		UpdatedAt:           base.UpdatedAt.Time,
		Source:              base.Source.String,
		BaseServings:        int(base.Servings),
		Servings:            int(base.Servings),
		ServingsDescription: base.ServingsDescription,
	}
	_ = base.WorkingTime.AssignTo(&res.WorkingTime)
	_ = base.WaitingTime.AssignTo(&res.WaitingTime)

	ingredients, err := app.Queries.GetTotalIngredientsForRecipe(ctx, int64(id))
	if err != nil {
		return nil, fmt.Errorf("querying ingredients for recipe %d: %w", id, err)
	}

	for _, ingredient := range ingredients {
		res.Ingredients = append(res.Ingredients, Ingredient{
			Name:     ingredient.Name,
			Amount:   float64(ingredient.TotalAmount),
			UnitName: ingredient.UnitName.String,
		})
	}

	steps, err := app.Queries.GetStepsForRecipeByID(ctx, int64(id))
	if err != nil {
		return nil, fmt.Errorf("querying steps for recipe %d: %w", id, err)
	}

	for _, step := range steps {
		s := Step{
			ID:          int(step.ID),
			RecipeID:    res.ID,
			Instruction: step.Instruction,
		}
		_ = step.StepTime.AssignTo(&s.Time)

		if step.Ingredients.Status == pgtype.Present {
			if err := step.Ingredients.AssignTo(&s.Ingredients); err != nil {
				return nil, fmt.Errorf("scanning ingredients into struct: %w", err)
			}
		}

		// Because we might have empty ingredients, filter those out.
		// TODO: can we get rid of them in the SQL query?
		var ingredients []Ingredient
		for _, i := range s.Ingredients {
			if i.Name != "" {
				ingredients = append(ingredients, i)
			}
		}
		s.Ingredients = ingredients

		res.Steps = append(res.Steps, s)
	}

	return res, nil
}

// UpdateRecipe updates basic information about a recipe. This includes:
//   - Name
//   - Servings
//   - Description
//
// All other properties are unaffected.
func (app *Application) UpdateRecipe(ctx context.Context, r *Recipe) error {
	err := app.Queries.UpdateBasicRecipeInformation(ctx, queries.UpdateBasicRecipeInformationParams{
		ID:                  int64(r.ID),
		Name:                r.Name,
		Servings:            int32(r.Servings),
		Description:         r.Description,
		ServingsDescription: r.ServingsDescription,
	})
	if err != nil {
		return fmt.Errorf("updating recipe %d in database: %w", r.ID, err)
	}

	return nil
}

type ListEntry struct {
	ID   int
	Name string
}

type Recipe struct {
	ID                  int
	Name                string
	Description         string
	WorkingTime         time.Duration
	WaitingTime         time.Duration
	CreatedAt           time.Time
	UpdatedAt           time.Time
	Source              string
	BaseServings        int // The number of servings that the recipe was written for.
	Servings            int // The current amount of servings, calculated with WithServings.
	ServingsDescription string

	Ingredients []Ingredient
	Steps       []Step
}

// WithServings recalculates the recipe with the given amount of servings.
// Any value less or equal to zero is ignored.
func (r *Recipe) WithServings(servings int) {
	if r.BaseServings == 0 {
		panic(fmt.Sprintf("base servings not set on recipe %d", r.ID))
	}

	if servings <= 0 || (r.BaseServings == servings) {
		return
	}

	baseServings, newServings := float64(r.BaseServings), float64(servings)
	r.Servings = servings

	for i, ingredient := range r.Ingredients {
		ingredient.Amount = ingredient.Amount / baseServings * newServings
		r.Ingredients[i] = ingredient
	}
	for i, step := range r.Steps {
		for j, ingredient := range step.Ingredients {
			ingredient.Amount = ingredient.Amount / baseServings * newServings
			step.Ingredients[j] = ingredient
		}
		r.Steps[i] = step
	}
}

type Ingredient struct {
	ID       int
	Name     string
	Amount   float64
	Note     string
	UnitName string

	// only used for the template "ingredient" and its delete button
	RecipeID int
	StepID   int
}
type Step struct {
	ID          int
	RecipeID    int
	Instruction string
	Time        time.Duration
	Ingredients []Ingredient
}
type Unit struct {
	ID   int
	Name string
}
