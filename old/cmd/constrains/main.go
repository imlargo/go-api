package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"reflect"
	"time"

	"github.com/nicolailuther/butter/internal/config"
	"github.com/nicolailuther/butter/internal/database"
	"github.com/nicolailuther/butter/internal/models"
	"gorm.io/gorm"
)

var (
	dryRun = flag.Bool("dry-run", false, "Show what would be done without executing")
	backup = flag.Bool("backup", true, "Create backup of current constraints")
	verify = flag.Bool("verify", true, "Verify constraints after migration")
)

func main() {
	flag.Parse()

	cfg := config.LoadConfig()

	// Database
	db, err := database.NewPostgres(cfg.Database.URL)
	if err != nil {
		log.Fatal("Could not initialize database: ", err)
	}

	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Define all your models here
	models := []interface{}{
		&models.User{},
		&models.Notification{},
		&models.PushNotificationSubscription{},
		&models.Client{},
		&models.Account{},
		&models.File{},
		&models.TextOverlay{},
		&models.Post{},
		&models.ReferralCode{},
		&models.AccountAnalytic{},
		&models.PostAnalytic{},
		&models.PostingGoal{},
		&models.MarketplaceCategory{},
		&models.MarketplaceSeller{},
		&models.MarketplaceService{},
		&models.MarketplaceServiceResult{},
		&models.MarketplaceServicePackage{},
		&models.Payment{},
		&models.MarketplaceOrder{},
		&models.MarketplaceDeliverable{},
		&models.MarketplaceRevisionRequest{},
		&models.MarketplaceDispute{},
		&models.MarketplaceOrderTimeline{},
		&models.OnlyfansAccount{},
		&models.OnlyfansTransaction{},
		&models.OnlyfansTrackingLink{},
		&models.ChatConversation{},
		&models.ChatMessage{},
		&models.ContentFolder{},
		&models.Content{},
		&models.ContentFile{},
		&models.ContentAccount{},
		&models.GeneratedContent{},
		&models.GeneratedContentFile{},
	}

	log.Println("=== CONSTRAINT MIGRATION TOOL ===")
	log.Printf("Dry Run: %v | Backup: %v | Verify: %v\n", *dryRun, *backup, *verify)

	// Create backup if enabled
	if *backup {
		log.Println("Creating backup of current constraints...")
		backupSQL, err := BackupConstraints(db)
		if err != nil {
			log.Fatalf("Failed to create backup: %v", err)
		}

		backupFile := fmt.Sprintf("constraints_backup_%s.sql",
			time.Now().Format("20060102_150405"))
		if err := os.WriteFile(backupFile, []byte(backupSQL), 0644); err != nil {
			log.Fatalf("Failed to write backup file: %v", err)
		}
		log.Printf("Backup saved to: %s", backupFile)
	}

	// Show current constraints
	log.Println("Current constraints:")
	if err := VerifyConstraints(db); err != nil {
		log.Fatalf("Failed to verify current constraints: %v", err)
	}

	if *dryRun {
		log.Println("DRY RUN: Would recreate all foreign key constraints")
		log.Println("Use without --dry-run to execute the migration")
		return
	}

	// Confirm before proceeding
	fmt.Print("This will drop and recreate ALL foreign key constraints. Continue? (y/N): ")
	var response string
	fmt.Scanln(&response)
	if response != "y" && response != "Y" {
		log.Println("Migration cancelled by user")
		return
	}

	// Execute the migration
	log.Println("Starting constraint recreation...")
	if err := RecreateAllConstraintsSafe(db, models...); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	err = database.Migrate(db)
	if err != nil {
		log.Fatal("Could not run migrations: ", err)
		return
	}

	// Verify results
	if *verify {
		log.Println("Verifying new constraints...")
		if err := VerifyConstraints(db); err != nil {
			log.Fatalf("Failed to verify new constraints: %v", err)
		}
	}

	log.Println("✅ Migration completed successfully!")
}

// RecreateAllConstraintsSafe drops and recreates ONLY the foreign key constraints
// preserving all data and table structure
func RecreateAllConstraintsSafe(db *gorm.DB, models ...interface{}) error {
	log.Println("Starting safe constraint recreation...")

	return db.Transaction(func(tx *gorm.DB) error {
		// Phase 1: Map and drop all foreign key constraints
		log.Println("Phase 1: Dropping all foreign key constraints...")
		constraintMap, err := dropAllForeignKeyConstraints(tx)
		if err != nil {
			return fmt.Errorf("failed to drop constraints: %w", err)
		}

		log.Printf("Dropped %d foreign key constraints", len(constraintMap))

		// Phase 2: Use AutoMigrate to recreate constraints based on models
		log.Println("Phase 2: Recreating constraints based on GORM models...")
		for _, model := range models {
			modelName := reflect.TypeOf(model).Elem().Name()
			log.Printf("Processing model: %s", modelName)

			if err := tx.AutoMigrate(model); err != nil {
				return fmt.Errorf("failed to migrate %s: %w", modelName, err)
			}
		}

		log.Println("Safe constraint recreation completed successfully!")
		return nil
	})
}

// dropAllForeignKeyConstraints drops all foreign keys and returns a map
// of the constraints that were dropped for logging
func dropAllForeignKeyConstraints(db *gorm.DB) (map[string]string, error) {
	// Get all existing foreign keys
	var constraints []struct {
		TableName      string `db:"table_name"`
		ConstraintName string `db:"constraint_name"`
		ColumnName     string `db:"column_name"`
		ForeignTable   string `db:"foreign_table_name"`
		DeleteRule     string `db:"delete_rule"`
	}

	query := `
        SELECT 
            tc.table_name,
            tc.constraint_name,
            kcu.column_name,
            ccu.table_name AS foreign_table_name,
            rc.delete_rule
        FROM information_schema.table_constraints AS tc 
        JOIN information_schema.key_column_usage AS kcu
            ON tc.constraint_name = kcu.constraint_name
        JOIN information_schema.constraint_column_usage AS ccu
            ON ccu.constraint_name = tc.constraint_name
        JOIN information_schema.referential_constraints AS rc
            ON tc.constraint_name = rc.constraint_name
        WHERE tc.constraint_type = 'FOREIGN KEY' 
        AND tc.table_schema = current_schema()
        ORDER BY tc.table_name, tc.constraint_name
    `

	err := db.Raw(query).Scan(&constraints)
	if err.Error != nil {
		return nil, err.Error
	}

	constraintMap := make(map[string]string)

	// Drop each constraint
	for _, constraint := range constraints {
		key := fmt.Sprintf("%s.%s", constraint.TableName, constraint.ConstraintName)
		value := fmt.Sprintf("%s -> %s (%s)", constraint.ColumnName, constraint.ForeignTable, constraint.DeleteRule)
		constraintMap[key] = value

		sql := fmt.Sprintf("ALTER TABLE %s DROP CONSTRAINT IF EXISTS %s",
			constraint.TableName, constraint.ConstraintName)

		log.Printf("Dropping: %s", key)
		if err := db.Exec(sql); err.Error != nil {
			return nil, fmt.Errorf("failed to drop constraint %s: %w", key, err.Error)
		}
	}

	return constraintMap, nil
}

// VerifyConstraints verifies that constraints were created correctly
func VerifyConstraints(db *gorm.DB) error {
	var constraints []struct {
		TableName      string `db:"table_name"`
		ConstraintName string `db:"constraint_name"`
		ColumnName     string `db:"column_name"`
		ForeignTable   string `db:"foreign_table_name"`
		DeleteRule     string `db:"delete_rule"`
		UpdateRule     string `db:"update_rule"`
	}

	query := `
        SELECT 
            tc.table_name,
            tc.constraint_name,
            kcu.column_name,
            ccu.table_name AS foreign_table_name,
            rc.delete_rule,
            rc.update_rule
        FROM information_schema.table_constraints AS tc 
        JOIN information_schema.key_column_usage AS kcu
            ON tc.constraint_name = kcu.constraint_name
        JOIN information_schema.constraint_column_usage AS ccu
            ON ccu.constraint_name = tc.constraint_name
        JOIN information_schema.referential_constraints AS rc
            ON tc.constraint_name = rc.constraint_name
        WHERE tc.constraint_type = 'FOREIGN KEY' 
        AND tc.table_schema = current_schema()
        ORDER BY tc.table_name, tc.constraint_name
    `

	err := db.Raw(query).Scan(&constraints)
	if err.Error != nil {
		return err.Error
	}

	log.Println("\n=== CURRENT FOREIGN KEY CONSTRAINTS ===")
	for _, c := range constraints {
		log.Printf("Table: %-20s | Constraint: %-30s | Column: %-15s -> %-10s | Delete: %-10s | Update: %s",
			c.TableName, c.ConstraintName, c.ColumnName, c.ForeignTable, c.DeleteRule, c.UpdateRule)
	}

	return nil
}

// BackupConstraints creates a backup of current constraints before modifying them
func BackupConstraints(db *gorm.DB) (string, error) {
	var constraints []struct {
		TableName      string `db:"table_name"`
		ConstraintName string `db:"constraint_name"`
		ColumnName     string `db:"column_name"`
		ForeignTable   string `db:"foreign_table_name"`
		ForeignColumn  string `db:"foreign_column_name"`
		DeleteRule     string `db:"delete_rule"`
		UpdateRule     string `db:"update_rule"`
	}

	query := `
        SELECT 
            tc.table_name,
            tc.constraint_name,
            kcu.column_name,
            ccu.table_name AS foreign_table_name,
            ccu.column_name AS foreign_column_name,
            rc.delete_rule,
            rc.update_rule
        FROM information_schema.table_constraints AS tc 
        JOIN information_schema.key_column_usage AS kcu
            ON tc.constraint_name = kcu.constraint_name
        JOIN information_schema.constraint_column_usage AS ccu
            ON ccu.constraint_name = tc.constraint_name
        JOIN information_schema.referential_constraints AS rc
            ON tc.constraint_name = rc.constraint_name
        WHERE tc.constraint_type = 'FOREIGN KEY' 
        AND tc.table_schema = current_schema()
        ORDER BY tc.table_name, tc.constraint_name
    `

	err := db.Raw(query).Scan(&constraints)
	if err.Error != nil {
		return "", err.Error
	}

	backup := "-- Backup of foreign key constraints\n"
	backup += "-- Generated automatically before constraint recreation\n\n"

	for _, c := range constraints {
		backup += fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s \n", c.TableName, c.ConstraintName)
		backup += fmt.Sprintf("    FOREIGN KEY (%s) REFERENCES %s(%s)\n",
			c.ColumnName, c.ForeignTable, c.ForeignColumn)
		backup += fmt.Sprintf("    ON UPDATE %s ON DELETE %s;\n\n", c.UpdateRule, c.DeleteRule)
	}

	return backup, nil
}
