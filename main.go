package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alecthomas/kong"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	s3manager "github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/coreos/go-systemd/v22/activation"
	"github.com/coreos/go-systemd/v22/daemon"
	"github.com/pin/tftp/v3"
)

type Args struct {
	S3uri url.URL `arg:"" required:"" name:"S3URI" help:"s3:// URI that identifies the target bucket and optional key prefix"`

	Region         string `name:"region" help:"AWS region where the bucket resides" placeholder:"REGION"`
	Retries        int    `short:"r" name:"retries" default:"5" help:"Number of retransmissions before the server disconnect the session"`
	Timeout        int    `short:"t" name:"timeout" default:"5000" help:"Timeout in milliseconds before the server retransmits a packet"`
	BlockSize      int    `short:"b" name:"blocksize" default:"512" help:"Maximum permitted block size in octets"`
	Anticipate     uint   `name:"anticipate" default:"0" help:"Size of anticipation window. Set 0 to disable sender anticipation (experimental)"`
	NoDualStack    bool   `name:"no-dualstack" help:"Disable S3 dualstack endpoint"`
	Accelerate     bool   `name:"accelerate" help:"Enable S3 Transfer Acceleration"`
	EndpointURL    string `name:"endpoint-url" help:"Use custom endpoint URL instead of default S3 endpoint" placeholder:"URL"`
	ForcePathStyle bool   `name:"force-path-style" help:"Use path-style URLs to access objects"`
	SinglePort     bool   `name:"single-port" help:"Serve all connections on a single UDP socket (experimental)"`
	Verbosity      int    `short:"v" name:"verbosity" default:"7" help:"Verbosity level for logging (0..8)"`
	DebugApi       bool   `name:"debug-api" env:"AWS_DEBUG" help:"Enable logging AWS API calls"`
}

type Config struct {
	Args

	bucket string
	prefix string
	s3     *s3.Client

	ctx context.Context
}

func (c *Config) awsOptions() (options []func(*awsConfig.LoadOptions) error) {
	if c.DebugApi {
		options = append(options, awsConfig.WithClientLogMode(
			aws.LogRequest|aws.LogRetries|aws.LogResponse,
		))
	}

	if c.Region != "" {
		options = append(options, awsConfig.WithRegion(c.Region))
	}

	if c.EndpointURL != "" {
		resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			if service == s3.ServiceID {
				return aws.Endpoint{
					PartitionID:       "aws",
					URL:               c.EndpointURL,
					SigningRegion:     c.Region,
					HostnameImmutable: true,
				}, nil
			}
			return aws.Endpoint{}, fmt.Errorf("unknown endpoint requested")
		})

		options = append(options, awsConfig.WithEndpointResolverWithOptions(resolver))
	}

	return
}

func (c *Config) logf(severity int, format string, args ...interface{}) (n int, error error) {
	return c.log(severity, fmt.Sprintf(format, args...))
}
func (c *Config) log(severity int, message interface{}) (n int, error error) {
	if severity >= c.Verbosity {
		return 0, nil
	}
	return fmt.Fprintf(os.Stderr, "<%d>%s\n", severity, message)
}

func (c *Config) handleRead(path string, rf io.ReaderFrom) error {
	xfer := rf.(tftp.OutgoingTransfer)
	remoteAddr := xfer.RemoteAddr()
	c.logf(6, "RRQ %s %s", remoteAddr.String(), path)

	key := prefixKey(c.prefix, path)
	c.logf(7, "GetObject %s %s", c.bucket, key)
	ret, err := c.s3.GetObject(c.ctx, &s3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return err
	}
	defer ret.Body.Close()

	tsize := ret.ContentLength
	c.logf(7, "%s tsize %d", remoteAddr.String(), tsize)
	xfer.SetSize(tsize)

	rf.ReadFrom(ret.Body)

	return nil
}

func buffer(wt io.WriterTo) io.Reader {
	buffer := bytes.NewBuffer(make([]byte, 0))
	wt.WriteTo(buffer)
	return buffer
}

func (c *Config) handleWrite(path string, wt io.WriterTo) error {
	xfer := wt.(tftp.IncomingTransfer)
	remoteAddr := xfer.RemoteAddr()
	c.logf(6, "WRQ %s %s", remoteAddr.String(), path)

	key := prefixKey(c.prefix, path)
	c.logf(7, "PutObject %s %s", c.bucket, key)
	_, err := s3manager.NewUploader(c.s3).Upload(c.ctx, &s3.PutObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
		Body:   buffer(wt),
	})
	if err != nil {
		return err
	}

	return nil
}

type hook struct {
	*Config
}

func (h hook) OnSuccess(result tftp.TransferStats) {
	addr := net.UDPAddr{IP: result.RemoteAddr, Port: result.Tid}
	h.logf(6, "FIN %s %s", addr.String(), result.Filename)
	h.logf(7, "stats %v", result)
}
func (h hook) OnFailure(result tftp.TransferStats, err error) {
	addr := net.UDPAddr{IP: result.RemoteAddr, Port: result.Tid}
	h.logf(4, "ERR %s %s: %s", addr.String(), result.Filename, err.Error())
	h.logf(7, "stats %v", result)
}

func (c *Config) hooks() hook {
	return hook{c}
}

func getConn() (net.PacketConn, error) {
	conns, err := activation.PacketConns()
	if err != nil {
		return nil, err
	}
	if len(conns) < 1 {
		return nil, errors.New("No datagram socket passed by system manager")
	}
	for _, c := range conns[1:] {
		c.Close()
	}
	return conns[0], nil
}

func parseArgs(ctx context.Context) (config Config, err error) {
	config.ctx = ctx

	parser, err := kong.New(&config.Args)
	if err != nil {
		panic(err)
	}

	parser.Model.HelpFlag.Short = 'h'

	_, err = parser.Parse(os.Args[1:])
	if err != nil {
		return
	}

	config.bucket, config.prefix, err = parseS3uri(config.S3uri)
	if err != nil {
		return
	}

	return
}

func main() {
	config, err := parseArgs(context.Background())
	if err != nil {
		config.Verbosity = 7
		config.log(2, err)
		os.Exit(1)
	}

	if sz := config.BlockSize; sz != 0 {
		if sz < 512 || sz > 65464 {
			config.log(2, "Block size is out of range (512..65464).")
			os.Exit(1)
		}
	}

	awsConfig, err := awsConfig.LoadDefaultConfig(config.ctx, config.awsOptions()...)
	if err != nil {
		config.log(2, err)
		os.Exit(1)
	}

	config.s3 = s3.NewFromConfig(awsConfig, func(o *s3.Options) {
		o.UsePathStyle = config.ForcePathStyle
		o.UseAccelerate = config.Accelerate
		o.UseDualstack = !config.NoDualStack
	})

	conn, err := getConn()
	if err != nil {
		config.log(2, err)
		os.Exit(1)
	}
	config.logf(5, "Listening on %s", conn.LocalAddr().String())

	server := tftp.NewServer(config.handleRead, config.handleWrite)
	server.SetTimeout(time.Duration(config.Timeout) * time.Millisecond)
	server.SetRetries(config.Retries)
	server.SetBlockSize(config.BlockSize)
	server.SetAnticipate(config.Anticipate)
	server.SetHook(config.hooks())
	if config.SinglePort {
		server.EnableSinglePort()
	}

	sigch := make(chan os.Signal, 1)
	signal.Notify(
		sigch,
		syscall.SIGQUIT,
	)
	go func() {
		for {
			switch <-sigch {
			case syscall.SIGQUIT:
				config.log(5, "Gracefully stopping server")
				daemon.SdNotify(false, daemon.SdNotifyStopping)
				server.Shutdown()
			}
		}
	}()

	config.log(5, "Starting server")
	daemon.SdNotify(false, daemon.SdNotifyReady)
	server.Serve(conn)
	if err != nil {
		config.log(2, err)
		os.Exit(1)
	}
}
