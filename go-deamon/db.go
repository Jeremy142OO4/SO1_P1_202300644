package main

import (
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

var (
	db     *sql.DB
	dbOnce sync.Once
)

// Ruta relativa desde go-deamon/
const dbPath = "../dashboard/data/metrics.db"

func InitDB() error {
	var err error
	dbOnce.Do(func() {
		db, err = sql.Open("sqlite", dbPath)
		if err != nil {
			return
		}
		if err = db.Ping(); err != nil {
			return
		}

		// PRAGMAs recomendados (no dañan si ya están en la DB)
		_, _ = db.Exec(`PRAGMA journal_mode=WAL;`)
		_, _ = db.Exec(`PRAGMA synchronous=NORMAL;`)

		// Opcional: validar que existan las tablas clave
		if err = ensureTableExists("lotes"); err != nil {
			return
		}
		if err = ensureTableExists("procesos_snapshot"); err != nil {
			return
		}
		if err = ensureTableExists("contenedores_snapshot"); err != nil {
			return
		}

		log.Printf("DB lista: %s\n", dbPath)
	})
	return err
}

func ensureTableExists(name string) error {
	var found string
	err := db.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name=?`, name).Scan(&found)
	if err == sql.ErrNoRows {
		return fmt.Errorf("no existe la tabla requerida: %s (revisa metrics.db)", name)
	}
	return err
}

func CloseDB() {
	if db != nil {
		_ = db.Close()
	}
}

// Crea un lote y devuelve id_lote
func CrearLote() (int64, error) {
	// tu tabla usa ts_utc TEXT NOT NULL
	ts := time.Now().UTC().Format(time.RFC3339Nano)

	res, err := db.Exec(`INSERT INTO lotes (ts_utc) VALUES (?)`, ts)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// Inserta snapshot de procesos (módulo 1)
func InsertarProcesosSnapshot(idLote int64, procs []Process) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	stmt, err := tx.Prepare(`
		INSERT OR REPLACE INTO procesos_snapshot
		(id_lote, pid, nombre, cmdline, vsz_kb, rss_kb, porcentaje_ram, porcentaje_cpu, utime, stime)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, p := range procs {
		_, err = stmt.Exec(
			idLote,
			p.PID,
			p.Name,
			p.Cmdline,
			p.VSZ,
			p.RSS,
			p.MemoryUsage,
			p.CPUUsage,
			p.UTime,
			p.STime,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}


type ContainerSnapshot struct {
	ContainerID string `json:"ContainerID"`
	CgroupPath  string `json:"CgroupPath"`
	RSSKB       int64  `json:"RSS_KB"`
	CPUJiffies  int64  `json:"CPU_Jiffies"`
	Procs       int64  `json:"Procs"`
}

func ResetDB() error {
	_, err := db.Exec(`
		DELETE FROM procesos_snapshot;
		DELETE FROM contenedores_snapshot;
		DELETE FROM lotes;
		DELETE FROM sqlite_sequence WHERE name IN ('lotes');
	`)
	return err
}
func InsertarContenedoresSnapshot(idLote int64, ci *ContInfo) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	stmt, err := tx.Prepare(`
		INSERT OR REPLACE INTO contenedores_snapshot
		(id_lote, id_contenedor, ruta_cgroup, rss_kb, cpu_jiffies, procesos)
		VALUES (?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, c := range ci.Containers {
		_, err = stmt.Exec(
			idLote,
			c.ContainerID,
			c.CgroupPath,
			int64(c.RSSKB),
			int64(c.CPUJiffies),
			int64(c.Procs),
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}