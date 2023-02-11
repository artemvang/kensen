package kensen

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/artemvang/kensen/internal/engines"
)

type MigrationStatusEnum uint8

const (
	Applied MigrationStatusEnum = iota
	Skipped
	Errored
)

func (m MigrationStatusEnum) String() string {
	switch m {
	case Applied:
		return "applied"
	case Skipped:
		return "skipped"
	case Errored:
		return "errored"
	}

	return ""
}

type MigrationStatus struct {
	Migration string
	Status    MigrationStatusEnum
	Err       error
}

type Kensen struct {
	engine         engines.Engine
	migrationsPath string
}

const migrationsTable string = "kensen"

func Create(uri string, migrationsPath string) (*Kensen, error) {
	var err error
	var engine engines.Engine

	if uri == "" {
		return nil, fmt.Errorf("database URI is not set")
	}
	if _, err := os.Stat(migrationsPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("migrations path directory does not exist")
	}

	protocol := strings.Split(uri, ":")[0]
	switch protocol {
	case "file":
		engine, err = engines.CreateSQLite(uri)
	case "postgresql":
		engine, err = engines.CreatePGSQL(uri)
	case "mysql":
		engine, err = engines.CreateMySQL(uri)
	}
	if err != nil {
		return nil, err
	}
	return &Kensen{engine: engine, migrationsPath: migrationsPath}, nil
}

func (k *Kensen) New(name string) (*string, error) {
	var migrationFile string

	date := time.Now().UTC().Format("2006-01-02")
	migrationFile = fmt.Sprintf("%s-%s.sql", date, name)

	f, err := os.Create(filepath.Join(k.migrationsPath, migrationFile))
	if err != nil {
		return nil, err
	}

	defer f.Close()

	_, err = f.WriteString("-- your sql code")
	if err != nil {
		return nil, err
	}

	return &migrationFile, nil
}

func (k *Kensen) Init() error {
	query := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (name varchar(255) NOT NULL UNIQUE)", migrationsTable)
	return k.engine.Execute(query)
}

func (k *Kensen) getApplied() ([]string, error) {
	query := fmt.Sprintf("SELECT name FROM %s ORDER BY name ASC", migrationsTable)
	rows, err := k.engine.Query(query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	migrations := make([]string, 0)

	for rows.Next() {
		var mig string
		if err := rows.Scan(&mig); err != nil {
			return nil, err
		}

		migrations = append(migrations, mig)
	}

	return migrations, nil
}

func (k *Kensen) getAvailable() ([]string, error) {
	files, err := os.ReadDir(k.migrationsPath)

	if err != nil {
		return nil, err
	}

	migrations := make([]string, 0)

	for _, f := range files {
		name := f.Name()
		if f.IsDir() || !strings.HasSuffix(name, ".sql") {
			continue
		}
		migrations = append(migrations, strings.TrimSuffix(name, ".sql"))
	}

	return migrations, nil
}

func (k *Kensen) applyMigration(migration string) MigrationStatus {
	sql, err := os.ReadFile(filepath.Join(k.migrationsPath, migration+".sql"))
	if err != nil {
		err = fmt.Errorf("error reading migration %s: %w", migration, err)
		return MigrationStatus{migration, Errored, err}
	}

	code := string(sql)

	err = k.engine.WithTx(func() error {
		return k.engine.Execute(code)
	})

	if err != nil {
		err = fmt.Errorf("error executing migration %s: %w", migration, err)
		return MigrationStatus{migration, Errored, err}
	}

	query := fmt.Sprintf("INSERT INTO %s(name) VALUES ($1)", migrationsTable)
	err = k.engine.Execute(query, migration)
	if err != nil {
		err = fmt.Errorf("error inserting migration %s: %w", migration, err)
		return MigrationStatus{migration, Errored, err}
	}

	return MigrationStatus{migration, Applied, nil}
}

func (k *Kensen) Apply() ([]MigrationStatus, error) {
	available, err := k.getAvailable()
	if err != nil {
		return nil, err
	}
	applied, err := k.getApplied()

	if err != nil {
		return nil, err
	}

	appliedSet := make(map[string]bool)
	for _, mig := range applied {
		appliedSet[mig] = true
	}

	migrationStatuses := make([]MigrationStatus, len(available))
	for i, migration := range available {
		if _, ok := appliedSet[migration]; ok {
			migrationStatuses[i] = MigrationStatus{migration, Skipped, nil}
			continue
		}

		migrationStatuses[i] = k.applyMigration(migration)
		if migrationStatuses[i].Status == Errored {
			break
		}
	}
	return migrationStatuses, nil
}
