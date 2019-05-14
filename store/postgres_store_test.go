// Copyright (c) 2019 Coinbase, Inc. See LICENSE

package store

import (
	"database/sql"
	"testing"

	"github.com/CoinbaseWallet/walletlinkd/config"
	_ "github.com/lib/pq" // register postgres adapter
	"github.com/stretchr/testify/require"
)

var db *sql.DB

func setup(t *testing.T) {
	var err error
	db, err = sql.Open("postgres", config.PostgresURL)
	require.Nil(t, err)

	_, err = db.Query("TRUNCATE store CASCADE")
	require.Nil(t, err)
}

func teardown() {
	db.Close()
	db = nil
}

func TestPostgresStoreNonExisting(t *testing.T) {
	setup(t)
	defer teardown()

	ps, err := NewPostgresStore(config.PostgresURL, "store")
	require.Nil(t, err)
	defer ps.Close()

	foo := dummy{}
	ok, err := ps.Get("foo", &foo)
	require.False(t, ok)
	require.Nil(t, err)
}

func TestPostgresStoreGetSet(t *testing.T) {
	setup(t)
	defer teardown()

	ps, err := NewPostgresStore(config.PostgresURL, "store")
	require.Nil(t, err)
	defer ps.Close()

	foo := dummy{X: 10, Y: 20}
	err = ps.Set("foo", &foo)
	require.Nil(t, err)

	bar := dummy{X: 111, Y: 222}
	err = ps.Set("bar", &bar)
	require.Nil(t, err)

	loadedFoo := dummy{}
	ok, err := ps.Get("foo", &loadedFoo)
	require.True(t, ok)
	require.Nil(t, err)

	require.Equal(t, 10, loadedFoo.X)
	require.Equal(t, 20, loadedFoo.Y)

	var j string
	err = db.QueryRow("SELECT value FROM store WHERE key = $1", "foo").Scan(&j)
	require.Nil(t, err)
	require.JSONEq(t, `{"x": 10, "y": 20}`, j)

	loadedBar := dummy{}
	ok, err = ps.Get("bar", &loadedBar)
	require.True(t, ok)
	require.Nil(t, err)

	require.Equal(t, 111, loadedBar.X)
	require.Equal(t, 222, loadedBar.Y)

	j = ""
	err = db.QueryRow("SELECT value FROM store WHERE key = $1", "bar").Scan(&j)
	require.Nil(t, err)
	require.JSONEq(t, `{"x": 111, "y": 222}`, j)
}

func TestPosgresStoreOverwrite(t *testing.T) {
	setup(t)
	defer teardown()

	ps, err := NewPostgresStore(config.PostgresURL, "store")
	require.Nil(t, err)
	defer ps.Close()

	foo := dummy{X: 10, Y: 20}
	err = ps.Set("foo", &foo)
	require.Nil(t, err)

	newFoo := dummy{X: 123, Y: 456}
	err = ps.Set("foo", &newFoo)
	require.Nil(t, err)

	loadedFoo := dummy{}
	ok, err := ps.Get("foo", &loadedFoo)
	require.True(t, ok)
	require.Nil(t, err)

	require.Equal(t, 123, loadedFoo.X)
	require.Equal(t, 456, loadedFoo.Y)

	var j string
	err = db.QueryRow("SELECT value FROM store WHERE key = $1", "foo").Scan(&j)
	require.Nil(t, err)
	require.JSONEq(t, `{"x": 123, "y": 456}`, j)
}

func TestPostgresStoreRemove(t *testing.T) {
	setup(t)
	defer teardown()

	ps, err := NewPostgresStore(config.PostgresURL, "store")
	require.Nil(t, err)
	defer ps.Close()

	foo := dummy{X: 10, Y: 20}
	err = ps.Set("foo", &foo)
	require.Nil(t, err)

	err = ps.Remove("foo")
	require.Nil(t, err)

	loadedFoo := dummy{}
	ok, err := ps.Get("foo", &loadedFoo)
	require.False(t, ok)
	require.Nil(t, err)

	n := -1
	err = db.QueryRow("SELECT count(*) FROM store WHERE key = $1", "foo").Scan(&n)
	require.Nil(t, err)
	require.Equal(t, 0, n)
}