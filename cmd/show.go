package cmd

import (
	"fmt"
	"strconv"

	"github.com/djcp/gorecipes/internal/db"
	"github.com/djcp/gorecipes/internal/ui"
	"github.com/spf13/cobra"
)

var showCmd = &cobra.Command{
	Use:   "show <id>",
	Short: "Display a recipe by ID",
	Args:  cobra.ExactArgs(1),
	RunE:  runShow,
}

func runShow(_ *cobra.Command, args []string) error {
	id, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil || id <= 0 {
		return fmt.Errorf("invalid recipe ID: %q", args[0])
	}

	recipe, err := db.GetRecipe(sqlDB, id)
	if err != nil {
		return fmt.Errorf("recipe not found: %w", err)
	}

	goHome, goAdd, deleteConfirmed, searchQuery, err := ui.RunDetailUI(recipe)
	if err != nil {
		return err
	}

	if deleteConfirmed {
		if err := db.DeleteRecipe(sqlDB, recipe.ID); err != nil {
			return fmt.Errorf("deleting recipe: %w", err)
		}
		return runList(nil, nil)
	}

	if goAdd {
		return runAdd(nil, nil)
	}

	if goHome {
		// User chose "home" from the detail view — open the interactive list,
		// carrying over any search query they typed in the detail view's search bar.
		listQuery = searchQuery
		return runList(nil, nil)
	}

	return nil
}
