package db

import (
	"context"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"testing"
	"url_shortener/pkg/short"
)

func TestDB_Insert(t *testing.T) {
	_sh := short.New()

	_db, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer func() { _ = _db.Close() }()

	originalURL := "original"
	shortURL := _sh.Short(originalURL)

	mock.ExpectBegin()
	mock.
		ExpectExec("INSERT INTO url_db").
		WithArgs(originalURL, shortURL).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	rows := sqlmock.NewRows([]string{"original_url"}).AddRow(originalURL)
	mock.
		ExpectQuery("SELECT original_url FROM url_db WHERE").
		WithArgs(shortURL).
		WillReturnRows(rows)

	db := DB{db: _db}

	err = db.Add(context.Background(), Row{OriginalURL: originalURL, ShortURL: shortURL})
	assert.Nil(t, err)

	_, err = db.GetOriginalURL(context.Background(), shortURL)
	assert.Nil(t, err)

	err = mock.ExpectationsWereMet()
	assert.Nil(t, err)
}

func TestDB_GetOriginalURLNoShortURL(t *testing.T) {
	_db, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer func() { _ = _db.Close() }()

	shortURL := "no short url"

	mock.
		ExpectQuery("SELECT original_url FROM url_db WHERE").
		WithArgs(shortURL).
		WillReturnRows(sqlmock.NewRows([]string{"original_url"}))

	db := DB{db: _db}

	_, err = db.GetOriginalURL(context.Background(), shortURL)
	assert.NotNil(t, err)

	err = mock.ExpectationsWereMet()
	assert.Nil(t, err)
}
