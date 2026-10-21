package migrations

import (
	"database/sql"

	"github.com/wibiesana/padi_go_core/database"
	"github.com/wibiesana/padi_go_core/migrator"
)

func init() {
	migrator.Register(
		"2026_01_01_000002_create_password_resets_table",
		func(db *sql.DB) error {
			driver := database.GetDriver()
			var createSQL string

			if driver == "postgres" {
				createSQL = `
				CREATE TABLE IF NOT EXISTS password_resets (
					id SERIAL PRIMARY KEY,
					email VARCHAR(255) NOT NULL,
					token VARCHAR(255) NOT NULL,
					expires_at BIGINT NOT NULL,
					created_by INT NULL,
					created_at BIGINT NULL,
					FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL
				);
				CREATE INDEX IF NOT EXISTS idx_pw_email ON password_resets (email);
				CREATE INDEX IF NOT EXISTS idx_pw_token ON password_resets (token);
				CREATE INDEX IF NOT EXISTS idx_pw_expires_at ON password_resets (expires_at);
				CREATE INDEX IF NOT EXISTS idx_pw_created_by ON password_resets (created_by);
				`
			} else if driver == "mysql" {
				createSQL = `
				CREATE TABLE IF NOT EXISTS password_resets (
					id INT AUTO_INCREMENT PRIMARY KEY,
					email VARCHAR(255) NOT NULL,
					token VARCHAR(255) NOT NULL,
					expires_at BIGINT NOT NULL,
					created_by INT NULL,
					created_at BIGINT NULL,
					FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL,
					INDEX idx_email (email),
					INDEX idx_token (token),
					INDEX idx_expires_at (expires_at),
					INDEX idx_created_by (created_by)
				) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
				`
			} else {
				createSQL = `
				CREATE TABLE IF NOT EXISTS password_resets (
					id INTEGER PRIMARY KEY AUTOINCREMENT,
					email VARCHAR(255) NOT NULL,
					token VARCHAR(255) NOT NULL,
					expires_at INTEGER NOT NULL,
					created_by INTEGER NULL,
					created_at INTEGER DEFAULT (strftime('%s', 'now')),
					FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL
				);
				CREATE INDEX IF NOT EXISTS idx_pw_email ON password_resets (email);
				CREATE INDEX IF NOT EXISTS idx_pw_token ON password_resets (token);
				CREATE INDEX IF NOT EXISTS idx_pw_expires_at ON password_resets (expires_at);
				CREATE INDEX IF NOT EXISTS idx_pw_created_by ON password_resets (created_by);
				`
			}

			return migrator.ExecHelper(db, createSQL)
		},
		func(db *sql.DB) error {
			_, err := db.Exec("DROP TABLE IF EXISTS password_resets")
			return err
		},
	)
}
