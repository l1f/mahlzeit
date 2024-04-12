package routes

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/l1f/mahlzeit/internal/app"
	"github.com/l1f/mahlzeit/internal/http/htmx"
	"github.com/l1f/mahlzeit/internal/http/httpreq"
	"github.com/l1f/mahlzeit/web/templates/pages/recipes"
	"github.com/robfig/bind"
)

func (a appWrapper) getAllRecipes(w http.ResponseWriter, r *http.Request) error {
	allRecipes, err := a.app.GetAllRecipes(r.Context())
	if err != nil {
		return err
	}

	component := recipes.AllRecipes(allRecipes)
	return component.Render(context.TODO(), w)
}

func (a appWrapper) getSingleRecipe(w http.ResponseWriter, r *http.Request) error {
	id := httpreq.MustIDParam(r, "id")
	recipe, err := a.app.GetSingleRecipe(r.Context(), id)
	if err != nil {
		return err
	}

	if servingsParam := r.URL.Query().Get("servings"); servingsParam != "" {
		p, err := strconv.Atoi(servingsParam)
		if err != nil {
			return fmt.Errorf("Invalid 'servings' parameter: %w", err)
		}
		recipe.WithServings(p)
	}

	component := recipes.SingleRecipe(*recipe)
	return component.Render(context.TODO(), w)
}

func (a appWrapper) getEditSingleRecipe(w http.ResponseWriter, r *http.Request) error {
	id := httpreq.MustIDParam(r, "id")

	recipe, err := a.app.GetSingleRecipe(r.Context(), id)
	if err != nil {
		return err
	}

	component := recipes.EditRecipe(*recipe)
	return component.Render(context.TODO(), w)
}

func (a appWrapper) postEditSingleRecipe(w http.ResponseWriter, r *http.Request) error {
	id := httpreq.MustIDParam(r, "id")
	if err := r.ParseForm(); err != nil {
		return err
	}

	data := struct {
		Name                string
		Servings            int
		ServingsDescription string
		Description         string
	}{}
	if err := bind.Request(r).All(&data); err != nil {
		return err
	}

	if err := a.app.UpdateRecipe(r.Context(), &app.Recipe{
		ID:                  id,
		Name:                data.Name,
		Description:         data.Description,
		BaseServings:        data.Servings,
		Servings:            data.Servings,
		ServingsDescription: data.ServingsDescription,
	}); err != nil {
		return err
	}

	http.Redirect(w, r, "/recipes/"+strconv.Itoa(id), http.StatusFound)
	return nil
}

func (a appWrapper) getAddStepToRecipe(w http.ResponseWriter, r *http.Request) error {
	if !htmx.IsHTMXRequest(r) {
		panic("progressive enhancement is not implemented yet")
	}

	id := httpreq.MustIDParam(r, "id")

	component := recipes.EditSingleStep(app.Step{
		RecipeID: id,
	})
	return component.Render(context.TODO(), w)
}

func (a appWrapper) getSingleStep(w http.ResponseWriter, r *http.Request) error {
	stepID, _ := httpreq.IDParam(r, "stepID")

	// If we invoke this route with step 0, we assume that the step hasn't been persisted yet.
	// In this case, we return nothing, so that the HTML node is removed again.
	if stepID == 0 {
		w.WriteHeader(http.StatusOK)
		return nil
	}

	step, err := a.app.GetStepByID(r.Context(), stepID)
	if err != nil {
		return err
	}

	component := recipes.SingleStep(step)
	return component.Render(context.TODO(), w)
}

func (a appWrapper) setStepToEditMode(w http.ResponseWriter, r *http.Request) error {
	if !htmx.IsHTMXRequest(r) {
		panic("progressive enhancement is not implemented yet")
	}

	step, err := a.app.GetStepByID(r.Context(), httpreq.MustIDParam(r, "stepID"))
	if err != nil {
		return err
	}

	component := recipes.EditSingleStep(step)
	return component.Render(context.TODO(), w)
}

func (a appWrapper) updateRecipeStep(w http.ResponseWriter, r *http.Request) error {
	recipeID := httpreq.MustIDParam(r, "id")
	stepID, _ := httpreq.IDParam(r, "stepID") // optional, because we differentiate between both states below
	if err := r.ParseForm(); err != nil {
		return err
	}

	data := struct {
		Instruction string
		Time        string
	}{}

	if err := bind.Request(r).Field(&data.Instruction, "instruction"); err != nil {
		return err
	}

	fmt.Println(data.Instruction))

	if err := bind.Request(r).Field(&data.Time, "time"); err != nil {
		return err
	}

	dur, _ := time.ParseDuration(data.Time)
	step := &app.Step{
		ID:          stepID,
		RecipeID:    recipeID,
		Instruction: data.Instruction,
		Time:        dur,
	}

	if stepID != 0 {
		if err := a.app.UpdateStep(r.Context(), *step); err != nil {
			return fmt.Errorf("updating step %d: %w", stepID, err)
		}
	} else {
		if err := a.app.AddStepToRecipe(r.Context(), recipeID, step); err != nil {
			return fmt.Errorf("adding step to recipe %d: %w", recipeID, err)
		}
	}

	if htmx.IsHTMXRequest(r) {
		component := recipes.SingleStep(*step)
		return component.Render(context.TODO(), w)
	}

	// if it not a HTMX request:
	http.Redirect(w, r, "/recipes/"+chi.URLParam(r, "id"), http.StatusNotFound)

	return nil
}

func (a appWrapper) deleteRecipeStep(w http.ResponseWriter, r *http.Request) error {
	id := httpreq.MustIDParam(r, "stepID")
	if err := a.app.DeleteRecipeStepByID(r.Context(), id); err != nil {
		return err
	}

	w.WriteHeader(http.StatusOK)
	return nil
}

func (a appWrapper) postAddNewRecipeStepIngredient(w http.ResponseWriter, r *http.Request) error {
	data := recipes.NewIngredientData{}

	ingredients, err := a.app.GetAllIngredients(r.Context())
	if err != nil {
		return err
	}

	units, err := a.app.GetAllUnits(r.Context())
	if err != nil {
		return err
	}

	data.Ingredients = ingredients
	data.Units = units

	stepID, err := httpreq.StrictIDParam(r, "stepID")
	if err != nil {
		return err
	}
	recipeID, err := httpreq.StrictIDParam(r, "id")
	if err != nil {
		return err
	}

	data.RecipeID = recipeID
	data.StepID = stepID

	if htmx.IsHTMXRequest(r) {
		component := recipes.NewIngredient(data, uuid.New().String())
		return component.Render(context.TODO(), w)
	}

	panic("progressive enhancement not yet implemented")
}

func (a appWrapper) postAddRecipeStepIngredient(w http.ResponseWriter, r *http.Request) error {
	recipeID := httpreq.MustIDParam(r, "id")
	stepID, err := httpreq.StrictIDParam(r, "stepID")
	if err != nil {
		return err
	}

	if err := r.ParseForm(); err != nil {
		return err
	}

	params := app.AddIngredientToStepParams{
		StepID:       stepID,
		IngredientID: parseIntWithDefault(r.PostFormValue("Ingredient")),
		Amount:       float64(parseIntWithDefault(r.PostFormValue("Amount"))),
		Note:         r.PostFormValue("Note"),
	}

	if unit := parseIntWithDefault(r.PostFormValue("Unit")); unit > 0 {
		params.UnitID = &unit
	}

	if err := a.app.AddIngredientToStep(r.Context(), params); err != nil {
		return err
	}

	ingredient, err := a.app.GetIngredient(r.Context(), params.IngredientID)
	if err != nil {
		return err
	}

	if htmx.IsHTMXRequest(r) {
		component := recipes.Ingredient(
			app.Ingredient{
				Name:     ingredient.Name,
				Amount:   params.Amount,
				Note:     params.Note,
				StepID:   stepID,
				RecipeID: recipeID,
			},
		)

		return component.Render(context.TODO(), w)
	}

	panic("progressive enhancement not yet implemented")

}

func (a appWrapper) deleteRecipeStepIngredient(_ http.ResponseWriter, r *http.Request) error {
	stepID, err := httpreq.StrictIDParam(r, "stepID")
	if err != nil {
		return err
	}

	ingredientID, err := httpreq.StrictIDParam(r, "ingredientID")
	if err != nil {
		return err
	}

	if err := a.app.DeleteIngredientFromStep(r.Context(), app.DeleteIngredientFromStepParams{
		StepID:       stepID,
		IngredientID: ingredientID,
	}); err != nil {
		return err
	}

	return nil
}

func parseIntWithDefault(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return i
}
