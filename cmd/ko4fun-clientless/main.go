package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"
	"text/tabwriter"
	"time"

	"github.com/0xhandle/ko4fun-clientless/internal/login"
	"github.com/0xhandle/ko4fun-clientless/internal/serverlist"
)

const (
	defaultHost = "login.ko4.fun"
	defaultPort = 15100
)

type options struct {
	host       string
	port       int
	username   string
	password   string
	jsonOutput bool
	timeout    time.Duration
}

func main() {
	opts := parseFlags()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if err := run(ctx, opts); err != nil {
		fmt.Fprintln(os.Stderr, "hata:", err)
		os.Exit(1)
	}
}

func parseFlags() options {
	var o options
	flag.StringVar(&o.username, "user", "", "kullanıcı adı (zorunlu)")
	flag.StringVar(&o.password, "password", "", "şifre (zorunlu)")
	flag.StringVar(&o.host, "host", defaultHost, "login sunucusu adresi")
	flag.IntVar(&o.port, "port", defaultPort, "login sunucusu portu")
	flag.BoolVar(&o.jsonOutput, "json", false, "sunucu listesini json olarak yazdır")
	flag.DurationVar(&o.timeout, "timeout", 10*time.Second, "bağlantı zaman aşımı")
	flag.Parse()
	return o
}

func run(ctx context.Context, o options) error {
	if o.username == "" || o.password == "" {
		flag.Usage()
		return errors.New("-user ve -password parametreleri zorunlu")
	}

	servers, err := login.FetchServerList(ctx, login.Config{
		Address:  net.JoinHostPort(o.host, strconv.Itoa(o.port)),
		Username: o.username,
		Password: o.password,
		Timeout:  o.timeout,
	})
	if err != nil {
		return err
	}

	if o.jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(servers)
	}
	return printTable(servers)
}

func printTable(servers []serverlist.Server) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tSUNUCU\tKANAL\tADRES\tOYUNCU\tKAPASİTE\tDURUM")
	for _, s := range servers {
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%d\t%d\t%d\n", s.ID, s.Name, s.Channel, s.Host, s.UserCount, s.MaxUsers, s.Status)
	}
	return w.Flush()
}
