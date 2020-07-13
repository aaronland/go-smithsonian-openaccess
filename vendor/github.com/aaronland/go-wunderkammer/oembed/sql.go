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
