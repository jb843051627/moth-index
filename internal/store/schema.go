package store

import "context"

const schema = `
CREATE TABLE IF NOT EXISTS stations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    code TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    habitat TEXT NOT NULL DEFAULT '',
    timezone TEXT NOT NULL,
    status TEXT NOT NULL,
    created_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS traps (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    station_id INTEGER NOT NULL REFERENCES stations(id),
    label TEXT NOT NULL,
    kind TEXT NOT NULL,
    status TEXT NOT NULL,
    installed_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS batches (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    station_id INTEGER NOT NULL REFERENCES stations(id),
    trap_id INTEGER NOT NULL REFERENCES traps(id),
    started_at TEXT NOT NULL,
    ended_at TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL,
    weather_note TEXT NOT NULL DEFAULT ''
);
CREATE TABLE IF NOT EXISTS specimens (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    batch_id INTEGER NOT NULL REFERENCES batches(id),
    tag TEXT NOT NULL UNIQUE,
    family TEXT NOT NULL DEFAULT '',
    genus TEXT NOT NULL DEFAULT '',
    species TEXT NOT NULL DEFAULT '',
    sex TEXT NOT NULL DEFAULT '',
    count INTEGER NOT NULL,
    status TEXT NOT NULL,
    notes TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS readings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    batch_id INTEGER NOT NULL REFERENCES batches(id),
    observed_at TEXT NOT NULL,
    temperature REAL NOT NULL,
    humidity REAL NOT NULL,
    lux REAL NOT NULL,
    rainfall REAL NOT NULL,
    source TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS taxa (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    family TEXT NOT NULL,
    genus TEXT NOT NULL,
    species TEXT NOT NULL,
    common_name TEXT NOT NULL DEFAULT '',
    authority TEXT NOT NULL DEFAULT '',
    UNIQUE(family, genus, species)
);
CREATE TABLE IF NOT EXISTS reviews (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    specimen_id INTEGER NOT NULL REFERENCES specimens(id),
    reviewer TEXT NOT NULL,
    decision TEXT NOT NULL,
    confidence REAL NOT NULL,
    note TEXT NOT NULL DEFAULT '',
    reviewed_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS signals (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    batch_id INTEGER NOT NULL REFERENCES batches(id),
    code TEXT NOT NULL,
    severity TEXT NOT NULL,
    message TEXT NOT NULL,
    active INTEGER NOT NULL,
    created_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS review_tasks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    kind TEXT NOT NULL,
    ref_id INTEGER NOT NULL,
    status TEXT NOT NULL,
    attempts INTEGER NOT NULL,
    error_text TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS audit_events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    entity TEXT NOT NULL,
    entity_id INTEGER NOT NULL,
    action TEXT NOT NULL,
    detail TEXT NOT NULL,
    created_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_traps_station ON traps(station_id);
CREATE INDEX IF NOT EXISTS idx_batches_station_time ON batches(station_id, started_at);
CREATE INDEX IF NOT EXISTS idx_specimens_batch ON specimens(batch_id);
CREATE INDEX IF NOT EXISTS idx_readings_batch_time ON readings(batch_id, observed_at);
CREATE INDEX IF NOT EXISTS idx_reviews_specimen ON reviews(specimen_id);
CREATE INDEX IF NOT EXISTS idx_signals_batch_active ON signals(batch_id, active);
CREATE INDEX IF NOT EXISTS idx_tasks_status_time ON review_tasks(status, created_at);
`

func (d *DB) EnsureSchema(ctx context.Context) error {
	_, err := d.Exec(ctx, schema)
	return err
}
