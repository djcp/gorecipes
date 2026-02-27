package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/djcp/gorecipes/internal/db"
	"github.com/djcp/gorecipes/internal/ui"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var (
	listQuery  string
	listStatus string
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Browse and search recipes",
	Long: `Open the interactive recipe browser.

Use / to search, arrow keys to navigate, enter to view a recipe.`,
	RunE: runList,
}

func init() {
	listCmd.Flags().StringVarP(&listQuery, "query", "q", "", "Filter by name or ingredient (non-interactive)")
	listCmd.Flags().StringVarP(&listStatus, "status", "s", "", "Filter by status (published, review, draft, etc.)")
}

func runList(_ *cobra.Command, _ []string) error {
	filter := db.RecipeFilter{
		Query:        listQuery,
		StatusFilter: listStatus,
	}

	// Non-interactive path: flag query set, or stdout is not a TTY.
	if listQuery != "" || !isTerminal() {
		recipes, err := db.ListRecipes(sqlDB, filter)
		if err != nil {
			return fmt.Errorf("loading recipes: %w", err)
		}
		if len(recipes) == 0 {
			fmt.Println(ui.MutedStyle.Render("\n  No recipes found."))
			fmt.Println(ui.MutedStyle.Render("  Add one with: gorecipes add <url>"))
			fmt.Println()
			return nil
		}
		fmt.Printf("\n  Found %d recipe(s):\n\n", len(recipes))
		for _, r := range recipes {
			courses := strings.Join(r.TagsByContext("courses"), ", ")
			fmt.Printf("  %3d  %-40s  %s\n", r.ID, r.Name, ui.MutedStyle.Render(courses))
		}
		fmt.Println()
		return nil
	}

	// Interactive path: loop between the list browser and recipe detail view.
	// When the user selects "home" from a recipe, control returns here and the
	// list re-opens, optionally with a search query carried over from the detail view.
	for {
		recipes, err := db.ListRecipes(sqlDB, filter)
		if err != nil {
			return fmt.Errorf("loading recipes: %w", err)
		}

		if len(recipes) == 0 {
			fmt.Println(ui.MutedStyle.Render("\n  No recipes found."))
			fmt.Println(ui.MutedStyle.Render("  Add one with: gorecipes add <url>"))
			fmt.Println()
			return nil
		}

		selectedID, goAdd, err := ui.RunListUI(recipes)
		if err != nil {
			return err
		}
		if goAdd {
			if err := runAdd(nil, nil); err != nil {
				return err
			}
			filter.Query = ""
			continue
		}
		if selectedID == 0 {
			// User quit from the list.
			break
		}

		recipe, err := db.GetRecipe(sqlDB, selectedID)
		if err != nil {
			return err
		}

		goHome, goAdd, searchQuery, err := ui.RunDetailUI(recipe)
		if err != nil {
			return err
		}
		if goAdd {
			if err := runAdd(nil, nil); err != nil {
				return err
			}
			filter.Query = ""
			continue
		}
		if !goHome {
			// User quit from the detail view.
			break
		}

		// User chose "home" — loop back to the list, applying any search they typed.
		filter.Query = searchQuery
	}

	return nil
}

func isTerminal() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}
