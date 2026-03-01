package cmd

import (
	"fmt"

	"github.com/djcp/gorecipes/internal/config"
	"github.com/djcp/gorecipes/internal/db"
	"github.com/djcp/gorecipes/internal/models"
	"github.com/djcp/gorecipes/internal/ui"
)

// runConfigUI opens the interactive config editor and saves changes if the user confirms.
func runConfigUI() error {
	configPath, _ := config.FilePath()
	saved, err := ui.RunConfigUI(cfg, configPath)
	if err != nil {
		return err
	}
	if saved {
		if err := cfg.Save(); err != nil {
			return fmt.Errorf("saving config: %w", err)
		}
	}
	return nil
}

func loadEditData() (ui.EditData, error) {
	ingNames, err := db.AllIngredientNames(sqlDB)
	if err != nil {
		return ui.EditData{}, err
	}
	units, err := db.AllUnits(sqlDB)
	if err != nil {
		return ui.EditData{}, err
	}
	tags := make(map[string][]string)
	for _, ctx := range models.AllTagContexts {
		names, err := db.AllTagsByContext(sqlDB, ctx)
		if err != nil {
			return ui.EditData{}, err
		}
		tags[ctx] = names
	}
	return ui.EditData{TagsByContext: tags, IngredientNames: ingNames, Units: units}, nil
}
