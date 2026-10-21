package migrations

import (
	"database/sql"

	"github.com/wibiesana/padi_go_core/database"
	"github.com/wibiesana/padi_go_core/migrator"
)

func init() {
	migrator.Register(
		"2026_01_01_000001_create_users_table",
		func(db *sql.DB) error {
			driver := database.GetDriver()
			var createSQL string

			if driver == "postgres" {
				createSQL = `
				CREATE TABLE IF NOT EXISTS users (
					id SERIAL PRIMARY KEY,
					name VARCHAR(255) NULL,
					username VARCHAR(50) UNIQUE,
					email VARCHAR(255) NOT NULL UNIQUE,
					password VARCHAR(255) NOT NULL,
					role VARCHAR(50) DEFAULT 'user',
					status VARCHAR(20) DEFAULT 'active',
					email_verified_at BIGINT NULL,
					remember_token VARCHAR(100) NULL,
					last_login_at BIGINT NULL,
					created_by INTEGER NULL,
					updated_by INTEGER NULL,
					created_at BIGINT NULL,
					updated_at BIGINT NULL,
					FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL,
					FOREIGN KEY (updated_by) REFERENCES users(id) ON DELETE SET NULL
				);
				CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
				CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
				CREATE INDEX IF NOT EXISTS idx_users_status ON users(status);
				CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);
				CREATE INDEX IF NOT EXISTS idx_users_created_by ON users(created_by);
				CREATE INDEX IF NOT EXISTS idx_users_updated_by ON users(updated_by);
				`
			} else if driver == "mysql" {
				createSQL = `
				CREATE TABLE IF NOT EXISTS users (
					id INT AUTO_INCREMENT PRIMARY KEY,
					name VARCHAR(255) NULL,
					username VARCHAR(50) UNIQUE,
					email VARCHAR(255) NOT NULL UNIQUE,
					password VARCHAR(255) NOT NULL,
					role VARCHAR(50) DEFAULT 'user',
					status VARCHAR(20) DEFAULT 'active',
					email_verified_at BIGINT NULL,
					remember_token VARCHAR(100) NULL,
					last_login_at BIGINT NULL,
					created_by INT NULL,
					updated_by INT NULL,
					created_at BIGINT NULL,
					updated_at BIGINT NULL,
					INDEX idx_users_email (email),
					INDEX idx_users_username (username),
					INDEX idx_users_status (status),
					INDEX idx_users_role (role),
					INDEX idx_users_created_by (created_by),
					INDEX idx_users_updated_by (updated_by),
					FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL,
					FOREIGN KEY (updated_by) REFERENCES users(id) ON DELETE SET NULL
				) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
				`
			} else {
				createSQL = `
				CREATE TABLE IF NOT EXISTS users (
					id INTEGER PRIMARY KEY AUTOINCREMENT,
					name VARCHAR(255) NULL,
					username VARCHAR(50) UNIQUE,
					email VARCHAR(255) NOT NULL UNIQUE,
					password VARCHAR(255) NOT NULL,
					role VARCHAR(50) DEFAULT 'user',
					status VARCHAR(20) DEFAULT 'active',
					email_verified_at INTEGER NULL,
					remember_token VARCHAR(100) NULL,
					last_login_at INTEGER NULL,
					created_by INTEGER NULL,
					updated_by INTEGER NULL,
					created_at INTEGER DEFAULT (strftime('%s', 'now')),
					updated_at INTEGER DEFAULT (strftime('%s', 'now')),
					FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL,
					FOREIGN KEY (updated_by) REFERENCES users(id) ON DELETE SET NULL
				);
				CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
				CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
				CREATE INDEX IF NOT EXISTS idx_users_status ON users(status);
				CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);
				CREATE INDEX IF NOT EXISTS idx_users_created_by ON users(created_by);
				CREATE INDEX IF NOT EXISTS idx_users_updated_by ON users(updated_by);
				`
			}

			return migrator.ExecHelper(db, createSQL)
		},
		func(db *sql.DB) error {
			_, err := db.Exec("DROP TABLE IF EXISTS users")
			return err
		},
	)
}
