package app

import (
	"context"
	"fmt"
)

func (app *Application) GetAllUnits(ctx context.Context) ([]Unit, error) {
	units, err := app.Queries.GetAllUnits(ctx)
	if err != nil {
		return nil, fmt.Errorf("querying all units: %w", err)
	}

	var res []Unit
	for _, u := range units {
		res = append(res, Unit{
			ID:   int(u.ID),
			Name: u.Name,
		})
	}

	return res, nil
}

func (app *Application) AddUnit(ctx context.Context, name string) (*int64, error) {
	id, err := app.Queries.AddUnit(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("error creating unit %s: %w", name, err)
	}

	return &id, nil
}

func (app *Application) DeleteUnit(ctx context.Context, unitId int64) error {
	err := app.Queries.DeleteUnit(ctx, unitId)
	if err != nil {
		return fmt.Errorf("error deleting unit with id %d: %w", unitId, err)
	}

	return nil
}
