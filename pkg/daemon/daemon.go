package daemon

import (
	"context"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"

	"url_shortener/pkg/config"
	"url_shortener/pkg/db"
	"url_shortener/pkg/server"
	"url_shortener/pkg/short"

	pb "url_shortener/pkg/grpc"

	log "github.com/sirupsen/logrus"
)

type Daemon struct {
	cfg config.Config

	ctx    context.Context
	cancel context.CancelFunc

	db         db.ShortenerDB
	grpcServer *grpc.Server
	urlServer  *server.Server
}

func New(ctx context.Context, cfg config.Config) (*Daemon, error) {
	d := &Daemon{cfg: cfg}
	var err error

	if ctx == nil {
		d.ctx, d.cancel = context.WithCancel(ctx)
	} else {
		d.ctx, d.cancel = context.WithCancel(context.Background())
	}

	d.db, err = db.New(ctx, cfg.DB)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	d.grpcServer = grpc.NewServer()

	d.urlServer, err = server.New(cfg.Server.LRUSize, d.db, short.New())
	if err != nil {
		log.Fatalf("cannot create URL server: %v", err)
	}

	return d, nil
}

func (d *Daemon) Run() error {
	log.Print("daemon started")

	serverErrC := make(chan error, 1)

	go func() {
		lis, err := net.Listen("tcp", d.cfg.Server.HostAddress())
		if err != nil {
			serverErrC <- err
			return
		}

		log.Printf("listening %s", d.cfg.Server.HostAddress())

		pb.RegisterURLShortenerServer(d.grpcServer, d.urlServer)

		serverErrC <- d.grpcServer.Serve(lis)
	}()

	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

		select {
		case <-d.ctx.Done():
			log.Print("interrupted by main context")
			return
		case <-stop:
			log.Print("interrupted by syscall")
			d.cancel()
		}
	}()

	select {
	case <-d.ctx.Done():
		d.ShutDown()
		return nil
	case err := <-serverErrC:
		d.ShutDown()
		return err
	}
}

func (d *Daemon) ShutDown() {
	defer d.cancel()

	d.grpcServer.GracefulStop()

	if err := d.db.Close(); err != nil {
		log.Printf("db closing error: %v", err)
	}

	log.Print("daemon is shut down")
}
