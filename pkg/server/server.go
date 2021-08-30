package server

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/golang-lru"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"url_shortener/pkg/db"
	"url_shortener/pkg/short"

	pb "url_shortener/pkg/grpc"

	log "github.com/sirupsen/logrus"
)

type Server struct {
	pb.UnimplementedURLShortenerServer

	db        db.ShortenerDB
	shortener short.Shortener

	// short -> original URL cache
	lruShortOrig *lru.Cache
	// original -> short URL cache
	lruOrigShort *lru.Cache
}

func New(lruSize int, db db.ShortenerDB, shortener short.Shortener) (*Server, error) {
	lruShortOrig, err := lru.New(lruSize)
	if err != nil {
		return nil, fmt.Errorf("server: cannot create lru: %w", err)
	}
	lruOrigShort, err := lru.New(lruSize)
	if err != nil {
		return nil, fmt.Errorf("server: cannot create lru: %w", err)
	}
	return &Server{
		db:           db,
		shortener:    shortener,
		lruOrigShort: lruOrigShort,
		lruShortOrig: lruShortOrig,
	}, nil
}

// Create shorts original URL and returns shorted
func (s *Server) Create(ctx context.Context, req *pb.CreateRequest) (*pb.CreateResponse, error) {
	if req.GetOriginalUrl() == "" {
		log.Debug("create: empty URL")
		return &pb.CreateResponse{}, status.Error(codes.Unknown, "cannot short empty URL")
	}

	// check not shorted
	isShort, err := s.isShort(ctx, req.GetOriginalUrl())
	if err != nil {
		log.Debugf("create: short check failed for URL=%s: %v:", req.GetOriginalUrl(), err)
		return &pb.CreateResponse{}, status.Error(codes.Unknown, "cannot short URL")
	}
	if isShort {
		log.Debugf("create: already shortened URL=%s", req.GetOriginalUrl())
		return &pb.CreateResponse{}, status.Error(codes.Unknown, "cannot short shortened URL")
	}

	shortURL, ok := s.lruOrigShort.Get(req.GetOriginalUrl())
	if ok {
		log.Debugf("create: original=%s short=%s (LRU)", req.GetOriginalUrl(), shortURL)
		return &pb.CreateResponse{ShortUrl: shortURL.(string)}, nil
	}

	return s.create(ctx, req)
}

// isShort checks if URL is shorted one
func (s *Server) isShort(ctx context.Context, url string) (bool, error) {
	// check cache
	_, ok := s.lruShortOrig.Get(url)
	if ok {
		return true, nil
	}

	// check db
	_, err := s.db.GetOriginalURL(ctx, url)
	if err != nil {
		if errors.Is(err, &db.NoRowError{}) {
			return false, nil
		}
		return false, fmt.Errorf("cannot check is URL short: %w", err)
	}
	return true, nil
}

// create adds new pair <original_url, short_url> to database
func (s *Server) create(ctx context.Context, req *pb.CreateRequest) (*pb.CreateResponse, error) {
	shortURL := s.shortener.Short(req.GetOriginalUrl())

	insertRow := db.Row{OriginalURL: req.GetOriginalUrl(), ShortURL: shortURL}
	if err := s.db.Add(ctx, insertRow); err != nil {
		log.Errorf("create: cannot add row original_url=%s: %v", req.GetOriginalUrl(), err)
		return &pb.CreateResponse{}, status.Error(codes.Unknown, "cannot add row")
	}

	// until no database success insert we can't update cache
	s.lruOrigShort.Add(req.GetOriginalUrl(), shortURL)

	log.Debugf("create: original=%s short=%s (DB)", req.GetOriginalUrl(), shortURL)

	return &pb.CreateResponse{ShortUrl: shortURL}, nil
}

// Get returns original URL by corresponding short URL
func (s *Server) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
	if req.GetShortUrl() == "" {
		log.Debug("get: empty URL")
		return &pb.GetResponse{}, status.Error(codes.Unknown, "empty short URL hasn't original URL")
	}

	originalURL, ok := s.lruShortOrig.Get(req.ShortUrl)
	if ok {
		log.Debugf("get: short=%s original=%s (LRU)", req.GetShortUrl(), originalURL)
		return &pb.GetResponse{OriginalUrl: originalURL.(string)}, nil
	}
	return s.get(ctx, req)
}

// get requests database for original URL
func (s *Server) get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
	originalURL, err := s.db.GetOriginalURL(ctx, req.GetShortUrl())
	if err != nil {
		if errors.Is(err, &db.NoRowError{}) {
			log.Debugf("get: no pair to provided short_url=%s", req.ShortUrl)
			return &pb.GetResponse{}, status.Error(codes.NotFound, "no pair to provided short URL")
		} else {
			log.Errorf("get: cannot get row with short_url=%s: %v", req.ShortUrl, err)
			return &pb.GetResponse{}, status.Error(codes.Unknown, "cannot get original URL")
		}
	}

	// until no database success select we can't update cache
	s.lruShortOrig.Add(req.GetShortUrl(), originalURL)

	log.Debugf("get: short=%s original=%s (DB)", req.GetShortUrl(), originalURL)

	return &pb.GetResponse{OriginalUrl: originalURL}, nil
}
