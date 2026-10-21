package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	_ "padi-template/database/migrations"
	"github.com/wibiesana/padi_go_core/config"
	"github.com/wibiesana/padi_go_core/database"
	"github.com/wibiesana/padi_go_core/generator"
	"github.com/wibiesana/padi_go_core/migrator"
	"github.com/wibiesana/padi_go_core/queue"
	"github.com/wibiesana/padi_go_core/wizard"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "padi",
	Short: "🌾 Padi REST API Framework CLI for Go",
	Long: `Padi REST API Framework - Industrial-Grade, Zero-Bloat API Engine for Go.
Rapidly build, migrate, and scaffold production-ready REST APIs.`,
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the Padi HTTP API server",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🚀 Starting Padi HTTP API server...")
		runCmd := exec.Command("go", "run", "main.go")
		runCmd.Stdout = os.Stdout
		runCmd.Stderr = os.Stderr
		runCmd.Stdin = os.Stdin
		if err := runCmd.Run(); err != nil {
			log.Fatalf("Server stopped: %v", err)
		}
	},
}

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Run pending database migrations",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		db, err := database.Connect(cfg)
		if err != nil {
			log.Fatalf("❌ Database connection error: %v", err)
		}

		if err := migrator.RunPending(db); err != nil {
			log.Fatalf("❌ Migration failed: %v", err)
		}
	},
}

var migrateRollbackCmd = &cobra.Command{
	Use:     "migrate:rollback",
	Aliases: []string{"rollback"},
	Short:   "Rollback the last batch of database migrations",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		db, err := database.Connect(cfg)
		if err != nil {
			log.Fatalf("❌ Database connection error: %v", err)
		}

		if err := migrator.RollbackLast(db); err != nil {
			log.Fatalf("❌ Rollback failed: %v", err)
		}
	},
}

var migrateStatusCmd = &cobra.Command{
	Use:   "migrate:status",
	Short: "Show status of all registered migrations",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		db, err := database.Connect(cfg)
		if err != nil {
			log.Fatalf("❌ Database connection error: %v", err)
		}

		statuses, err := migrator.Status(db)
		if err != nil {
			log.Fatalf("❌ Failed to fetch migration status: %v", err)
		}

		fmt.Println("---------------------------------------------------------------")
		fmt.Printf("%-50s | %-6s | %-5s\n", "Migration", "Ran?", "Batch")
		fmt.Println("---------------------------------------------------------------")
		for _, s := range statuses {
			ranStr := "No"
			if s.Ran {
				ranStr = "Yes"
			}
			batchStr := "-"
			if s.Batch > 0 {
				batchStr = fmt.Sprintf("%d", s.Batch)
			}
			fmt.Printf("%-50s | %-6s | %-5s\n", s.Name, ranStr, batchStr)
		}
		fmt.Println("---------------------------------------------------------------")
	},
}

var migrateFreshCmd = &cobra.Command{
	Use:   "migrate:fresh",
	Short: "Rollback all migrations and re-run all from start",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		db, err := database.Connect(cfg)
		if err != nil {
			log.Fatalf("❌ Database connection error: %v", err)
		}

		fmt.Println("🔄 Resetting database migrations...")
		if err := migrator.Fresh(db); err != nil {
			log.Fatalf("❌ Fresh migration failed: %v", err)
		}
		fmt.Println("✨ All migrations fresh and executed successfully!")
	},
}

var flagRealtime bool

func askRealtimePrompt(cmd *cobra.Command) bool {
	if cmd.Flags().Changed("realtime") {
		return flagRealtime
	}
	fmt.Print("📡 Enable Real-time SSE broadcasting hooks for this CRUD? [y/N]: ")
	var input string
	_, _ = fmt.Scanln(&input)
	input = strings.TrimSpace(strings.ToLower(input))
	return input == "y" || input == "yes"
}

var generateCrudCmd = &cobra.Command{
	Use:     "generate:crud [table_name]",
	Aliases: []string{"g"},
	Short:   "Generate Base Model, Custom Model, Controller, Resource, and Route for a database table",
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		tableName := args[0]
		cfg := config.Load()
		_, err := database.Connect(cfg)
		if err != nil {
			log.Fatalf("❌ Database connection error: %v", err)
		}

		enableRealtime := askRealtimePrompt(cmd)

		gen := generator.New(".").WithRealtime(enableRealtime)
		if err := gen.GenerateCRUD(tableName); err != nil {
			log.Fatalf("❌ CRUD Generation failed: %v", err)
		}
	},
}

var generateCrudAllCmd = &cobra.Command{
	Use:     "generate:crud-all",
	Aliases: []string{"ga"},
	Short:   "Generate CRUD for all tables in the database",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		_, err := database.Connect(cfg)
		if err != nil {
			log.Fatalf("❌ Database connection error: %v", err)
		}

		enableRealtime := askRealtimePrompt(cmd)

		gen := generator.New(".").WithRealtime(enableRealtime)
		if err := gen.GenerateAll(); err != nil {
			log.Fatalf("❌ CRUD Generation All failed: %v", err)
		}
	},
}

var queueWorkCmd = &cobra.Command{
	Use:     "queue:work [queue_name]",
	Aliases: []string{"queue"},
	Short:   "Start processing background queue jobs",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		_, err := database.Connect(cfg)
		if err != nil {
			log.Fatalf("❌ Database connection error: %v", err)
		}

		qName := "default"
		if len(args) > 0 && args[0] != "" {
			qName = args[0]
		}

		queue.Work(qName, 0)
	},
}

var makeModelCmd = &cobra.Command{
	Use:     "make:model [name]",
	Aliases: []string{"create:model"},
	Short:   "Create a new model file",
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		modelName := generator.TableNameToModelName(args[0])
		dir := "app/Models"
		_ = os.MkdirAll(dir, 0755)
		filePath := filepath.Join(dir, modelName+".go")

		if _, err := os.Stat(filePath); err == nil {
			log.Fatalf("Model %s already exists at %s", modelName, filePath)
		}

		template := fmt.Sprintf(`package models

import (
	"time"
)

type %s struct {
	ID        uint       `+"`db:\"id\" json:\"id\"`"+`
	CreatedAt *time.Time `+"`db:\"created_at\" json:\"created_at\"`"+`
	UpdatedAt *time.Time `+"`db:\"updated_at\" json:\"updated_at\"`"+`
}

func (%s) TableName() string {
	return "%s"
}
`, modelName, modelName, strings.ToLower(args[0]))

		if err := os.WriteFile(filePath, []byte(template), 0644); err != nil {
			log.Fatalf("Failed to create model: %v", err)
		}
		fmt.Printf("✓ Model created: %s\n", filePath)
	},
}

var makeControllerCmd = &cobra.Command{
	Use:     "make:controller [name]",
	Aliases: []string{"create:controller"},
	Short:   "Create a new controller file",
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctrlName := generator.TableNameToModelName(args[0])
		if !filepath.IsAbs(ctrlName) && len(ctrlName) > 10 && ctrlName[len(ctrlName)-10:] == "Controller" {
			ctrlName = ctrlName[:len(ctrlName)-10]
		}

		dir := "app/Controllers"
		_ = os.MkdirAll(dir, 0755)
		filePath := filepath.Join(dir, ctrlName+"Controller.go")

		if _, err := os.Stat(filePath); err == nil {
			log.Fatalf("Controller %s already exists at %s", ctrlName+"Controller", filePath)
		}

		template := fmt.Sprintf(`package controllers

import (
	"net/http"

	"github.com/wibiesana/padi_go_core/response"
)

type %sController struct{}

func New%sController() *%sController {
	return &%sController{}
}

// Index handles GET request for %s resource
func (c *%sController) Index(w http.ResponseWriter, r *http.Request) {
	response.Success(w, nil, "%sController index endpoint")
}
`, ctrlName, ctrlName, ctrlName, ctrlName, ctrlName, ctrlName, ctrlName)

		if err := os.WriteFile(filePath, []byte(template), 0644); err != nil {
			log.Fatalf("Failed to create controller: %v", err)
		}
		fmt.Printf("✓ Controller created: %s\n", filePath)
	},
}

var makeMigrationCmd = &cobra.Command{
	Use:     "make:migration [name]",
	Aliases: []string{"create:migration"},
	Short:   "Create a new migration file",
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		rawName := args[0]
		timestamp := time.Now().Format("2006_01_02_150405")
		migName := fmt.Sprintf("%s_%s", timestamp, rawName)

		dir := "database/migrations"
		_ = os.MkdirAll(dir, 0755)
		filePath := filepath.Join(dir, migName+".go")

		template := fmt.Sprintf(`package migrations

import (
	"database/sql"

	"github.com/wibiesana/padi_go_core/migrator"
)

func init() {
	migrator.Register(
		"%s",
		func(db *sql.DB) error {
			// Define Up migration logic (e.g. CREATE TABLE)
			return nil
		},
		func(db *sql.DB) error {
			// Define Down rollback logic (e.g. DROP TABLE)
			return nil
		},
	)
}
`, migName)

		if err := os.WriteFile(filePath, []byte(template), 0644); err != nil {
			log.Fatalf("Failed to create migration: %v", err)
		}
		fmt.Printf("✓ Migration created: %s\n", filePath)
	},
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update Padi Core dependencies and packages to their latest versions",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🌾 Updating Padi dependencies...")

		// 1. Run go get -u ./...
		getCmd := exec.Command("go", "get", "-u", "./...")
		getCmd.Stdout = os.Stdout
		getCmd.Stderr = os.Stderr
		if err := getCmd.Run(); err != nil {
			log.Fatalf("❌ Failed running 'go get -u ./...': %v", err)
		}

		// 2. Run go mod tidy
		tidyCmd := exec.Command("go", "mod", "tidy")
		tidyCmd.Stdout = os.Stdout
		tidyCmd.Stderr = os.Stderr
		if err := tidyCmd.Run(); err != nil {
			log.Fatalf("❌ Failed running 'go mod tidy': %v", err)
		}

		fmt.Println("✨ Padi dependencies updated successfully!")
	},
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize the application (Run Interactive Setup Wizard)",
	Run: func(cmd *cobra.Command, args []string) {
		wiz := wizard.New(".")
		if err := wiz.Run(); err != nil {
			log.Fatalf("Initialization failed: %v", err)
		}
	},
}

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build production binary named according to APP_NAME in .env",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		appName := "padi-app"
		if cfg != nil && cfg.AppName != "" {
			appName = strings.ToLower(strings.TrimSpace(cfg.AppName))
			appName = strings.ReplaceAll(appName, " ", "-")
			appName = strings.ReplaceAll(appName, "_", "-")
		}

		outputName := appName
		// If running on Windows, add .exe suffix
		if os.Getenv("GOOS") == "windows" || filepath.Separator == '\\' {
			outputName += ".exe"
		}

		fmt.Printf("🔨 Building production binary: %s ...\n", outputName)

		buildArgs := []string{"build", "-ldflags", "-s -w", "-o", outputName, "main.go"}
		execCmd := exec.Command("go", buildArgs...)
		execCmd.Stdout = os.Stdout
		execCmd.Stderr = os.Stderr

		if err := execCmd.Run(); err != nil {
			log.Fatalf("❌ Build failed: %v", err)
		}

		fmt.Printf("✨ Successfully built %s!\n", outputName)
	},
}

func init() {

	generateCrudCmd.Flags().BoolVarP(&flagRealtime, "realtime", "r", false, "Generate CRUD with real-time SSE broadcasting hooks in controllers")
	generateCrudAllCmd.Flags().BoolVarP(&flagRealtime, "realtime", "r", false, "Generate CRUD with real-time SSE broadcasting hooks in controllers")

	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(buildCmd)
	rootCmd.AddCommand(migrateCmd)
	rootCmd.AddCommand(migrateRollbackCmd)
	rootCmd.AddCommand(migrateStatusCmd)
	rootCmd.AddCommand(migrateFreshCmd)
	rootCmd.AddCommand(generateCrudCmd)
	rootCmd.AddCommand(generateCrudAllCmd)
	rootCmd.AddCommand(queueWorkCmd)
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(makeModelCmd)
	rootCmd.AddCommand(makeControllerCmd)
	rootCmd.AddCommand(makeMigrationCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
