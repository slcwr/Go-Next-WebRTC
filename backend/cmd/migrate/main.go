// cmd/migrate/main.go
package main

import (
	"database/sql"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

type Migration struct {
	Version   string
	Name      string
	SQL       string
	Timestamp time.Time
}

func main() {
	// コマンドラインフラグ
	var (
		command = flag.String("command", "up", "Command to run: up, down, status, create")
		name    = flag.String("name", "", "Name for new migration (used with create command)")
		steps   = flag.Int("steps", 0, "Number of migrations to run (0 = all)")
		dryRun  = flag.Bool("dry-run", false, "Show what would be executed without running")
	)
	flag.Parse()

	// .envファイルの読み込み
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// データベース接続
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		log.Fatal("DB_DSN environment variable is required")
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// マイグレーションテーブルの作成
	if err := createMigrationTable(db); err != nil {
		log.Fatalf("Failed to create migration table: %v", err)
	}

	// コマンド実行
	switch *command {
	case "up":
		if err := migrateUp(db, *steps, *dryRun); err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
	case "down":
		if err := migrateDown(db, *steps, *dryRun); err != nil {
			log.Fatalf("Rollback failed: %v", err)
		}
	case "status":
		if err := showStatus(db); err != nil {
			log.Fatalf("Failed to show status: %v", err)
		}
	case "create":
		if *name == "" {
			log.Fatal("Migration name is required for create command")
		}
		if err := createMigration(*name); err != nil {
			log.Fatalf("Failed to create migration: %v", err)
		}
	default:
		log.Fatalf("Unknown command: %s", *command)
	}
}

// createMigrationTable マイグレーション管理テーブルを作成
func createMigrationTable(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS schema_migrations (
		version VARCHAR(255) PRIMARY KEY,
		executed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`
	_, err := db.Exec(query)
	return err
}

// getMigrations マイグレーションファイルを取得
func getMigrations() ([]Migration, error) {
	migrationDir := "database/migrations"
	files, err := ioutil.ReadDir(migrationDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read migration directory: %w", err)
	}

	var migrations []Migration
	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".sql") {
			continue
		}

		// 特定のファイルをスキップ
		if strings.Contains(file.Name(), "init.sql") ||
			strings.Contains(file.Name(), "rollback.sql") ||
			strings.Contains(file.Name(), "cleanup.sql") ||
			strings.Contains(file.Name(), "seed_data.sql") {
			continue
		}

		content, err := ioutil.ReadFile(filepath.Join(migrationDir, file.Name()))
		if err != nil {
			return nil, fmt.Errorf("failed to read migration file %s: %w", file.Name(), err)
		}

		// ファイル名から番号を抽出（例: 001_create_users_table.sql）
		parts := strings.Split(file.Name(), "_")
		if len(parts) < 2 {
			continue
		}

		migrations = append(migrations, Migration{
			Version: parts[0],
			Name:    strings.TrimSuffix(file.Name(), ".sql"),
			SQL:     string(content),
		})
	}

	// バージョンでソート
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

// getExecutedMigrations 実行済みマイグレーションを取得
func getExecutedMigrations(db *sql.DB) (map[string]bool, error) {
	rows, err := db.Query("SELECT version FROM schema_migrations")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	executed := make(map[string]bool)
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		executed[version] = true
	}

	return executed, rows.Err()
}

// migrateUp マイグレーションを実行
func migrateUp(db *sql.DB, steps int, dryRun bool) error {
	migrations, err := getMigrations()
	if err != nil {
		return err
	}

	executed, err := getExecutedMigrations(db)
	if err != nil {
		return err
	}

	count := 0
	for _, migration := range migrations {
		if executed[migration.Version] {
			continue
		}

		if steps > 0 && count >= steps {
			break
		}

		log.Printf("Migrating %s...", migration.Name)

		if dryRun {
			log.Printf("Would execute:\n%s", migration.SQL)
			count++
			continue
		}

		// トランザクション開始
		tx, err := db.Begin()
		if err != nil {
			return err
		}

		// マイグレーション実行
		if _, err := tx.Exec(migration.SQL); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to execute migration %s: %w", migration.Name, err)
		}

		// 実行記録
		if _, err := tx.Exec("INSERT INTO schema_migrations (version) VALUES (?)", migration.Version); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to record migration %s: %w", migration.Name, err)
		}

		if err := tx.Commit(); err != nil {
			return err
		}

		log.Printf("Migrated %s successfully", migration.Name)
		count++
	}

	if count == 0 {
		log.Println("No migrations to run")
	} else {
		log.Printf("Migrated %d migrations", count)
	}

	return nil
}

// migrateDown ロールバック処理
func migrateDown(db *sql.DB, steps int, dryRun bool) error {
	if steps == 0 {
		steps = 1 // デフォルトは1つ戻る
	}

	// 実装簡略化のため、rollback.sqlを実行
	log.Printf("Rolling back %d migrations...", steps)
	
	if dryRun {
		log.Println("Would execute rollback")
		return nil
	}

	// 注意: 実際のプロジェクトでは、各マイグレーションに対応する
	// ロールバックスクリプトを用意すべき
	log.Println("Rollback functionality not fully implemented")
	log.Println("Please use rollback.sql manually if needed")

	return nil
}

// showStatus マイグレーションステータスを表示
func showStatus(db *sql.DB) error {
	migrations, err := getMigrations()
	if err != nil {
		return err
	}

	executed, err := getExecutedMigrations(db)
	if err != nil {
		return err
	}

	fmt.Println("Migration Status:")
	fmt.Println("-----------------")

	for _, migration := range migrations {
		status := "pending"
		if executed[migration.Version] {
			status = "executed"
		}
		fmt.Printf("%s | %s | %s\n", migration.Version, status, migration.Name)
	}

	return nil
}

// createMigration 新しいマイグレーションファイルを作成
func createMigration(name string) error {
	timestamp := time.Now().Format("20060102150405")
	version := fmt.Sprintf("%s", timestamp)
	
	// 既存のマイグレーション数を取得して連番を作成
	migrations, _ := getMigrations()
	number := fmt.Sprintf("%03d", len(migrations)+1)
	
	filename := fmt.Sprintf("database/migrations/%s_%s.sql", number, name)
	
	template := fmt.Sprintf(`-- Migration: %s
-- Created at: %s

-- Write your migration SQL here

-- Example:
-- CREATE TABLE IF NOT EXISTS table_name (
--     id BIGINT AUTO_INCREMENT PRIMARY KEY,
--     created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
-- );
`, name, time.Now().Format("2006-01-02 15:04:05"))

	if err := ioutil.WriteFile(filename, []byte(template), 0644); err != nil {
		return fmt.Errorf("failed to create migration file: %w", err)
	}

	log.Printf("Created migration: %s", filename)
	return nil
}

