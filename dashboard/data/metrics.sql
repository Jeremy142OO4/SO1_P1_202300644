PRAGMA journal_mode=WAL;
PRAGMA synchronous=NORMAL;


CREATE TABLE IF NOT EXISTS lotes (
  id_lote     INTEGER PRIMARY KEY AUTOINCREMENT,
  ts_utc      TEXT NOT NULL
);


CREATE TABLE IF NOT EXISTS procesos_snapshot (
  id_lote         INTEGER NOT NULL,
  pid             INTEGER NOT NULL,
  nombre          TEXT NOT NULL,
  cmdline         TEXT,
  vsz_kb          INTEGER,
  rss_kb          INTEGER,
  porcentaje_ram  REAL,
  porcentaje_cpu  REAL,
  utime           INTEGER,
  stime           INTEGER,
  PRIMARY KEY (id_lote, pid),
  FOREIGN KEY (id_lote) REFERENCES lotes(id_lote)
);

CREATE INDEX IF NOT EXISTS idx_proc_lote ON procesos_snapshot(id_lote);
CREATE INDEX IF NOT EXISTS idx_proc_lote_cpu ON procesos_snapshot(id_lote, porcentaje_cpu DESC);
CREATE INDEX IF NOT EXISTS idx_proc_lote_ram ON procesos_snapshot(id_lote, rss_kb DESC);


CREATE TABLE IF NOT EXISTS contenedores_snapshot (
  id_lote         INTEGER NOT NULL,
  id_contenedor   TEXT NOT NULL,
  ruta_cgroup     TEXT,
  rss_kb          INTEGER,
  cpu_jiffies     INTEGER,
  procesos        INTEGER,
  PRIMARY KEY (id_lote, id_contenedor),
  FOREIGN KEY (id_lote) REFERENCES lotes(id_lote)
);

CREATE INDEX IF NOT EXISTS idx_cont_lote ON contenedores_snapshot(id_lote);
CREATE INDEX IF NOT EXISTS idx_cont_lote_cpu ON contenedores_snapshot(id_lote, cpu_jiffies DESC);
CREATE INDEX IF NOT EXISTS idx_cont_lote_ram ON contenedores_snapshot(id_lote, rss_kb DESC);
