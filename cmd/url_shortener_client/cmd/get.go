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

func newGetCmd(address string) *cobra.Command {

	return &cobra.Command{
		Use:                "get shortURL",
		Short:              "Get original URL from given short URL",
		Args:               cobra.ExactArgs(1),
		FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
		Run: func(cmd *cobra.Command, args []string) {
			shortURL := args[0]

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
			resp, err := client.Get(ctx, &pb.GetRequest{ShortUrl: shortURL})
			if err != nil {
				log.Fatalf("cannot get original URL: %v", err)
			}
			fmt.Println(resp.GetOriginalUrl())
		},
	}

}
