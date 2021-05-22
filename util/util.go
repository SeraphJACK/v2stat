package util

import (
	"context"
	pb "github.com/SeraphJACK/v2stat/command"
	"google.golang.org/grpc"
	"strconv"
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
	defer conn.Close()
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

var units = map[int]string{
	0: "B",
	1: "KiB",
	2: "MiB",
	3: "GiB",
	4: "TiB",
}

func FormatTraffic(traffic int64) string {
	unit := 0
	num := float64(traffic)
	for unit <= 3 && num >= 1000 {
		unit++
		num /= 1024
	}
	return strconv.FormatFloat(num, 'f', 2, 64) + " " + units[unit]
}

func ThisMonth() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
}

func Today() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
}

func Day(offset int) time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day()+offset, 0, 0, 0, 0, now.Location())
}
