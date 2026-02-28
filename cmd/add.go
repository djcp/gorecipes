package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/djcp/gorecipes/internal/db"
	"github.com/djcp/gorecipes/internal/models"
	"github.com/djcp/gorecipes/internal/services"
	"github.com/djcp/gorecipes/internal/ui"
	"github.com/spf13/cobra"
)

var pasteFlag bool

var addCmd = &cobra.Command{
	Use:   "add [url]",
	Short: "Add a recipe from a URL or pasted text",
	Long: `Add a recipe using AI extraction.

Provide a URL as an argument, or use --paste to enter text directly:

  gorecipes add https://example.com/recipe
  gorecipes add --paste`,
	Args: cobra.MaximumNArgs(1),
	RunE: runAdd,
}

func init() {
	addCmd.Flags().BoolVarP(&pasteFlag, "paste", "p", false, "Paste recipe text instead of providing a URL")
}

func runAdd(_ *cobra.Command, args []string) error {
	// If a URL was provided as a CLI argument, validate it and skip the input UI.
	initialURL := ""
	if len(args) == 1 {
		initialURL = strings.TrimSpace(args[0])
		if !isURL(initialURL) {
			return fmt.Errorf("URL must start with http:// or https://")
		}
	}

	// launchFn creates the draft recipe and runs the extraction pipeline.
	// It is called by AddModel once the user submits the form.
	launchFn := ui.PipelineLaunchFn(func(ctx context.Context, sourceURL, sourceText string, onStep func(int, string)) (int64, error) {
		draft := &models.Recipe{
			Status:     models.StatusDraft,
			SourceURL:  sourceURL,
			SourceText: sourceText,
			Name:       "(importing...)",
		}
		recipeID, err := db.CreateRecipe(sqlDB, draft)
		if err != nil {
			return 0, fmt.Errorf("creating recipe: %w", err)
		}
		pipelineCfg := services.PipelineConfig{
			DB:     sqlDB,
			Client: services.NewAnthropicClient(cfg.AnthropicAPIKey),
			Model:  cfg.AnthropicModel,
			OnStep: onStep,
		}
		_, err = services.RunPipeline(ctx, pipelineCfg, recipeID)
		return recipeID, err
	})

	recipeID, goHome, goAdd, pipeErr := ui.RunAddUI(pasteFlag, initialURL, launchFn)

	if goAdd {
		return runAdd(nil, nil)
	}
	if goHome {
		return runList(nil, nil)
	}
	// pipeErr was already shown in the TUI; return nil so cobra doesn't double-print.
	if pipeErr != nil || recipeID <= 0 {
		return nil
	}

	// Pipeline succeeded — open the recipe in the interactive detail view.
	recipe, err := db.GetRecipe(sqlDB, recipeID)
	if err != nil {
		return err
	}
	goHome2, goAdd2, deleteConfirmed, searchQuery, err := ui.RunDetailUI(recipe)
	if err != nil {
		return err
	}
	if deleteConfirmed {
		if err := db.DeleteRecipe(sqlDB, recipe.ID); err != nil {
			return fmt.Errorf("deleting recipe: %w", err)
		}
		return runList(nil, nil)
	}
	if goAdd2 {
		return runAdd(nil, nil)
	}
	if goHome2 {
		listQuery = searchQuery
		return runList(nil, nil)
	}
	return nil
}

func isURL(s string) bool {
	return strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://")
}
