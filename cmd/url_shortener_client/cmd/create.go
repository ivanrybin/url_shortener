package cmd

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/spf13/cobra"

	"google.golang.org/grpc"

	pb "url_shortener/pkg/grpc"
)

func newCreateCmd(address string) *cobra.Command {

	return &cobra.Command{
		Use:                "create originalURL",
		Short:              "Create short URL from given original URL",
		Args:               cobra.ExactArgs(1),
		FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
		Run: func(cmd *cobra.Command, args []string) {
			originalURL := args[0]

			conn, err := grpc.Dial(
				address,
				grpc.WithInsecure(),
				grpc.WithBlock(),
				grpc.FailOnNonTempDialError(true),
			)
			if err != nil {
				log.Fatalf("cannot connect to `%s`: %v", address, err)
			}
			defer func() { _ = conn.Close() }()

			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			defer cancel()

			client := pb.NewURLShortenerClient(conn)
			resp, err := client.Create(ctx, &pb.CreateRequest{OriginalUrl: originalURL})
			if err != nil {
				log.Fatalf("cannot create short URL: %v", err)
			}
			fmt.Println(resp.GetShortUrl())
		},
	}

}
