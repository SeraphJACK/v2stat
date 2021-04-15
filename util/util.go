package util

import (
	"context"
	pb "github.com/SeraphJACK/v2stat/command"
	"google.golang.org/grpc"
	"strings"
	"time"
)

func QueryStats(addr string, pattern string, reset bool) (*pb.QueryStatsResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	conn, err := grpc.DialContext(ctx, addr, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, err
	}
	return queryStats(ctx, conn, pattern, reset)
}

func ExtractUser(name string) string {
	return strings.Split(name, ">>>")[1]
}

func queryStats(ctx context.Context, conn *grpc.ClientConn, pattern string, reset bool) (*pb.QueryStatsResponse, error) {
	r := &pb.QueryStatsRequest{
		Pattern: pattern,
		Reset_:  reset,
	}
	client := pb.NewStatsServiceClient(conn)
	res, err := client.QueryStats(ctx, r)
	if err != nil {
		return nil, err
	}
	return res, nil
}
