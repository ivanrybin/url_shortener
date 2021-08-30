package server

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
	"url_shortener/pkg/grpc"
	"url_shortener/pkg/short"
)
import "url_shortener/pkg/db"

type dbMock struct {
	originalShort map[string]string
	shortOriginal map[string]string
}

func NewDB() *dbMock {
	return &dbMock{originalShort: map[string]string{}, shortOriginal: map[string]string{}}
}

func (d *dbMock) Close() error { return nil }

func (d *dbMock) Add(_ context.Context, row db.Row) error {
	d.originalShort[row.OriginalURL] = row.ShortURL
	d.shortOriginal[row.ShortURL] = row.OriginalURL
	return nil
}

func (d *dbMock) GetOriginalURL(_ context.Context, shortURL string) (string, error) {
	originalURL, ok := d.shortOriginal[shortURL]
	if !ok {
		return "", &db.NoRowError{}
	}
	return originalURL, nil
}

func initAll(lruSize int) (*Server, *dbMock, short.Shortener, error) {
	_db := NewDB()
	_sh := short.New()
	_serv, err := New(lruSize, _db, _sh)
	return _serv, _db, _sh, err
}

func TestServer_Create(t *testing.T) {
	serv, _db, _sh, err := initAll(10)
	assert.Nil(t, err)

	for _, originalURL := range []string{
		"a", "b", "c", "d", "e", "f",
	} {
		shortURL := _sh.Short(originalURL)

		resp, err := serv.Create(context.Background(), &grpc.CreateRequest{OriginalUrl: originalURL})
		assert.Nil(t, err)
		assert.Equal(t, resp.ShortUrl, shortURL)

		shortDB, ok := _db.originalShort[originalURL]
		assert.True(t, ok)
		assert.Equal(t, shortURL, shortDB)

		originalDB, ok := _db.shortOriginal[shortURL]
		assert.True(t, ok)
		assert.Equal(t, originalURL, originalDB)
	}
}

func TestServer_CreateEmpty(t *testing.T) {
	serv, _, _, err := initAll(10)
	assert.Nil(t, err)

	_, err = serv.Create(context.Background(), &grpc.CreateRequest{OriginalUrl: ""})
	assert.NotNil(t, err)
}

func TestServer_Get(t *testing.T) {
	serv, _db, _sh, err := initAll(10)
	assert.Nil(t, err)

	urls := []string{"a", "b", "c", "d", "e", "f"}

	for _, originalURL := range urls {
		_db.originalShort[originalURL] = _sh.Short(originalURL)
		_db.shortOriginal[_sh.Short(originalURL)] = originalURL
	}

	for _, originalURL := range urls {
		shortURL := _sh.Short(originalURL)

		resp, err := serv.Get(context.Background(), &grpc.GetRequest{ShortUrl: shortURL})
		assert.Nil(t, err)
		assert.Equal(t, resp.GetOriginalUrl(), originalURL)
	}
}

func TestServer_GetNotExist(t *testing.T) {
	serv, _, _, err := initAll(10)
	assert.Nil(t, err)

	_, err = serv.Get(context.Background(), &grpc.GetRequest{ShortUrl: "not exist"})
	assert.NotNil(t, err)
}

func TestServer_GetEmpty(t *testing.T) {
	serv, _, _, err := initAll(10)
	assert.Nil(t, err)

	_, err = serv.Get(context.Background(), &grpc.GetRequest{ShortUrl: ""})
	assert.NotNil(t, err)
}
