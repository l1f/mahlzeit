package routes

import (
	"context"
	"errors"
	"net/http"

	"github.com/l1f/mahlzeit/internal/app"
	"github.com/l1f/mahlzeit/internal/http/httpreq"
	"github.com/l1f/mahlzeit/web/templates/pages/units"
)

func (a appWrapper) getAllUnits(w http.ResponseWriter, r *http.Request) error {
	allUnits, err := a.app.GetAllUnits(r.Context())
	if err != nil {
		return err
	}

	component := units.ListUnits(allUnits)
	return component.Render(context.TODO(), w)
}

func (a appWrapper) deleteUnit(w http.ResponseWriter, r *http.Request) error {
	unitId := httpreq.MustIDParam(r, "id")

	err := a.app.DeleteUnit(r.Context(), int64(unitId))
	if err != nil {
		return err
	}

	return nil
}

func (a appWrapper) postUnit(w http.ResponseWriter, r *http.Request) error {
	if err := r.ParseForm(); err != nil {
		return err
	}

	name := r.FormValue("name")
	if name == "" {
		return errors.New("field 'name' is empty")
	}

	id, err := a.app.AddUnit(r.Context(), name)
	if err != nil {
		return err
	}

	component := units.ListItem(app.Unit{
		ID:   int(*id),
		Name: name,
	})

	return component.Render(context.TODO(), w)
}
