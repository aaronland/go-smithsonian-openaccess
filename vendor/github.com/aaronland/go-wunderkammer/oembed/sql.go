package oembed

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/url"
)

type SQLOEmbedDatabase struct {
	OEmbedDatabase
	conn *sql.DB
}

func NewSQLOEmbedDatabase(ctx context.Context, uri string) (OEmbedDatabase, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, err
	}

	driver := u.Host
	dsn := u.Path

	if u.RawQuery != "" {
		dsn = fmt.Sprintf("%s?%s", dsn, u.RawQuery)
	}

	conn, err := sql.Open(driver, dsn)

	if err != nil {
		return nil, err
	}

	if driver == "sqlite3" {

		pragma := []string{
			"PRAGMA JOURNAL_MODE=OFF",
			"PRAGMA SYNCHRONOUS=OFF",
			"PRAGMA LOCKING_MODE=EXCLUSIVE",
			// https://www.gaia-gis.it/gaia-sins/spatialite-cookbook/html/system.html
			"PRAGMA PAGE_SIZE=4096",
			"PRAGMA CACHE_SIZE=1000000",
		}

		for _, p := range pragma {

			_, err = conn.Exec(p)

			if err != nil {
				return nil, err
			}
		}

	}

	db := &SQLOEmbedDatabase{
		conn: conn,
	}

	return db, nil
}

func (db *SQLOEmbedDatabase) Close() error {
	return db.conn.Close()
}

func (db *SQLOEmbedDatabase) AddOEmbed(ctx context.Context, rec *Photo) error {

	body, err := json.Marshal(rec)

	if err != nil {
		return err
	}

	url := rec.URL
	object_uri := rec.ObjectURI

	tx, err := db.conn.BeginTx(ctx, nil)

	if err != nil {
		return err
	}

	q := "INSERT OR REPLACE INTO oembed (url, object_uri, body) VALUES(?, ?, ?)"

	_, err = tx.ExecContext(ctx, q, url, object_uri, body)

	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (db *SQLOEmbedDatabase) GetRandomOEmbed(ctx context.Context) (*Photo, error) {

	q := "SELECT body FROM oembed ORDER BY RANDOM() LIMIT 1"

	row := db.conn.QueryRowContext(ctx, q)

	var body []byte

	err := row.Scan(&body)

	if err != nil {
		return nil, err
	}

	var rec *Photo

	err = json.Unmarshal(body, &rec)

	if err != nil {
		return nil, err
	}

	return rec, nil
}

func (db *SQLOEmbedDatabase) GetOEmbedWithURL(ctx context.Context, url string) (*Photo, error) {

	q := "SELECT body FROM oembed WHERE url = ?"

	row := db.conn.QueryRowContext(ctx, q, url)

	var body []byte

	err := row.Scan(&body)

	if err != nil {
		return nil, err
	}

	var rec *Photo

	err = json.Unmarshal(body, &rec)

	if err != nil {
		return nil, err
	}

	return rec, nil
}

func (db *SQLOEmbedDatabase) GetOEmbedWithObjectURI(ctx context.Context, object_uri string) ([]*Photo, error) {

	q := "SELECT body FROM oembed WHERE object_uri = ?"

	rows, err := db.conn.QueryContext(ctx, q, object_uri)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	photos := make([]*Photo, 0)

	for rows.Next() {

		var body []byte

		err := rows.Scan(&body)

		if err != nil {
			return nil, err
		}

		var ph *Photo

		err = json.Unmarshal(body, &ph)

		if err != nil {
			return nil, err
		}

		photos = append(photos, ph)
	}

	err = rows.Close()

	if err != nil {
		return nil, err
	}

	err = rows.Err()

	if err != nil {
		return nil, err
	}

	return photos, nil
}

func (db *SQLOEmbedDatabase) GetOEmbedWithCallback(ctx context.Context, cb OEmbedDatabaseCallback) error {

	q := "SELECT body FROM oembed"

	rows, err := db.conn.QueryContext(ctx, q)

	if err != nil {

		return err
	}

	defer rows.Close()

	for rows.Next() {

		var body []byte

		err := rows.Scan(&body)

		if err != nil {
			return err
		}

		var ph *Photo

		err = json.Unmarshal(body, &ph)

		if err != nil {
			return err
		}

		err = cb(ctx, ph)

		if err != nil {
			return err
		}
	}

	err = rows.Close()

	if err != nil {
		return err
	}

	err = rows.Err()

	if err != nil {
		return err
	}

	return nil
}
