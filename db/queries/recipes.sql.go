// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.17.0
// source: recipes.sql

package queries

import (
	"context"
	"database/sql"
	"time"

	"github.com/jackc/pgtype"
)

const addNewEmptyStep = `-- name: AddNewEmptyStep :one
insert into steps (recipe_id, sort_order, instruction, time)
select $1,
	   coalesce(max(sort_order), 0) + 1,
	   '',
	   '0 seconds'
from steps
where recipe_id = $1
returning steps.id, steps.recipe_id, steps.sort_order, steps.instruction, steps.time
`

func (q *Queries) AddNewEmptyStep(ctx context.Context, recipeID int64) (Step, error) {
	row := q.db.QueryRow(ctx, addNewEmptyStep, recipeID)
	var i Step
	err := row.Scan(
		&i.ID,
		&i.RecipeID,
		&i.SortOrder,
		&i.Instruction,
		&i.Time,
	)
	return i, err
}

const addRecipe = `-- name: AddRecipe :one
insert into recipes(name, description, working_time, waiting_time, created_at, updated_at, created_by, source, servings,
					servings_description)
values ($1,
		$2,
		$3,
		$4,
		now(),
		now(),
		$5,
		$6,
		$7,
		$8)
returning id, created_at
`

type AddRecipeParams struct {
	Name                string
	Description         string
	WorkingTime         pgtype.Interval
	WaitingTime         pgtype.Interval
	CreatedBy           int64
	Source              sql.NullString
	Servings            int32
	ServingsDescription string
}

type AddRecipeRow struct {
	ID        int64
	CreatedAt time.Time
}

func (q *Queries) AddRecipe(ctx context.Context, arg AddRecipeParams) (AddRecipeRow, error) {
	row := q.db.QueryRow(ctx, addRecipe,
		arg.Name,
		arg.Description,
		arg.WorkingTime,
		arg.WaitingTime,
		arg.CreatedBy,
		arg.Source,
		arg.Servings,
		arg.ServingsDescription,
	)
	var i AddRecipeRow
	err := row.Scan(&i.ID, &i.CreatedAt)
	return i, err
}

const deleteStepByID = `-- name: DeleteStepByID :exec
delete
from steps
where id = $1
`

func (q *Queries) DeleteStepByID(ctx context.Context, id int64) error {
	_, err := q.db.Exec(ctx, deleteStepByID, id)
	return err
}

const getAllRecipesByName = `-- name: GetAllRecipesByName :many
select id, name
from recipes
order by name
`

type GetAllRecipesByNameRow struct {
	ID   int64
	Name string
}

func (q *Queries) GetAllRecipesByName(ctx context.Context) ([]GetAllRecipesByNameRow, error) {
	rows, err := q.db.Query(ctx, getAllRecipesByName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetAllRecipesByNameRow
	for rows.Next() {
		var i GetAllRecipesByNameRow
		if err := rows.Scan(&i.ID, &i.Name); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getRecipeByID = `-- name: GetRecipeByID :one
select id,
	   name,
	   description,
	   working_time,
	   waiting_time,
	   created_at,
	   updated_at,
	   created_by,
	   source,
	   servings,
	   servings_description
from recipes
where id = $1
`

func (q *Queries) GetRecipeByID(ctx context.Context, id int64) (Recipe, error) {
	row := q.db.QueryRow(ctx, getRecipeByID, id)
	var i Recipe
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Description,
		&i.WorkingTime,
		&i.WaitingTime,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.CreatedBy,
		&i.Source,
		&i.Servings,
		&i.ServingsDescription,
	)
	return i, err
}

const getStepByID = `-- name: GetStepByID :one
select id, recipe_id, sort_order, instruction, time from steps where id = $1
`

func (q *Queries) GetStepByID(ctx context.Context, id int64) (Step, error) {
	row := q.db.QueryRow(ctx, getStepByID, id)
	var i Step
	err := row.Scan(
		&i.ID,
		&i.RecipeID,
		&i.SortOrder,
		&i.Instruction,
		&i.Time,
	)
	return i, err
}

const getStepsForRecipeByID = `-- name: GetStepsForRecipeByID :many
select steps.id,
	   instruction,
	   "time"  as step_time,
	   -- To get the ingredients within the same query and avoiding n+1 query pipelines,
	   -- those are built as a JSON object using jsonb_build_object.
	   -- Because the values can have NULL values due to the left join below, we strip those values
	   -- with jsonb_strip_nulls. And in the end, they are grouped inside an array with jsonb_agg.
	   jsonb_agg(jsonb_strip_nulls(jsonb_build_object(
			   'id', ingredients.id,
			   'stepID', steps.id,
			   'recipeID', steps.recipe_id,
			   'unitName', units.name,
			   'name', ingredients.name,
			   'amount', step_ingredients.amount,
			   'note', step_ingredients.note
		   ))) as ingredients
from steps
		 left join step_ingredients on steps.id = step_ingredients.step_id
		 left join ingredients on step_ingredients.ingredients_id = ingredients.id
		 left join units on units.id = step_ingredients.unit_id
where steps.recipe_id = $1
group by steps.id, steps.sort_order, "time", instruction
order by steps.sort_order
`

type GetStepsForRecipeByIDRow struct {
	ID          int64
	Instruction string
	StepTime    pgtype.Interval
	Ingredients pgtype.JSONB
}

func (q *Queries) GetStepsForRecipeByID(ctx context.Context, id int64) ([]GetStepsForRecipeByIDRow, error) {
	rows, err := q.db.Query(ctx, getStepsForRecipeByID, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetStepsForRecipeByIDRow
	for rows.Next() {
		var i GetStepsForRecipeByIDRow
		if err := rows.Scan(
			&i.ID,
			&i.Instruction,
			&i.StepTime,
			&i.Ingredients,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getTotalIngredientsForRecipe = `-- name: GetTotalIngredientsForRecipe :many
select ingredients.name,
	   units.name                   as unit_name,
	   sum(step_ingredients.amount) as total_amount
from steps
		 inner join step_ingredients on steps.id = step_ingredients.step_id
		 inner join ingredients on ingredients.id = step_ingredients.ingredients_id
		 left join units on units.id = step_ingredients.unit_id
where steps.recipe_id = $1
group by ingredients.name, units.name
order by ingredients.name, total_amount desc
`

type GetTotalIngredientsForRecipeRow struct {
	Name        string
	UnitName    sql.NullString
	TotalAmount int64
}

func (q *Queries) GetTotalIngredientsForRecipe(ctx context.Context, id int64) ([]GetTotalIngredientsForRecipeRow, error) {
	rows, err := q.db.Query(ctx, getTotalIngredientsForRecipe, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetTotalIngredientsForRecipeRow
	for rows.Next() {
		var i GetTotalIngredientsForRecipeRow
		if err := rows.Scan(&i.Name, &i.UnitName, &i.TotalAmount); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const updateBasicRecipeInformation = `-- name: UpdateBasicRecipeInformation :exec
update recipes
set name                 = $1,
	servings             = $2,
	description          = $3,
	servings_description = $4,
	updated_at           = now()
where id = $5
`

type UpdateBasicRecipeInformationParams struct {
	Name                string
	Servings            int32
	Description         string
	ServingsDescription string
	ID                  int64
}

func (q *Queries) UpdateBasicRecipeInformation(ctx context.Context, arg UpdateBasicRecipeInformationParams) error {
	_, err := q.db.Exec(ctx, updateBasicRecipeInformation,
		arg.Name,
		arg.Servings,
		arg.Description,
		arg.ServingsDescription,
		arg.ID,
	)
	return err
}

const updateStepByID = `-- name: UpdateStepByID :exec
update steps
set instruction = $1,
	time        = $2
where id = $3
`

type UpdateStepByIDParams struct {
	Instruction string
	Time        pgtype.Interval
	ID          int64
}

func (q *Queries) UpdateStepByID(ctx context.Context, arg UpdateStepByIDParams) error {
	_, err := q.db.Exec(ctx, updateStepByID, arg.Instruction, arg.Time, arg.ID)
	return err
}
