package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/defaults"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/coreos/go-systemd/v22/activation"
	"github.com/coreos/go-systemd/v22/daemon"
	"github.com/jessevdk/go-flags"
	"github.com/pin/tftp"
)

type Config struct {
	Args struct {
		S3uri string `positional-arg-name:"S3URI"`
	} `positional-args:"true" required:"true"`
	Verbosity   int  `short:"v" long:"verbosity" default:"7" description:"Verbosity level for logging (0..8)"`
	Timeout     int  `short:"t" long:"timeout" default:"5000" description:"Timeout in milliseconds before the server retransmits a packet"`
	Retries     int  `short:"r" long:"retries" default:"5" description:"Number of retransmissions before the server disconnect the session"`
	NoDualStack bool `long:"no-dualstack" description:"Disable S3 dualstack endpoint"`
	DebugApi    bool `long:"debug-api" env:"AWS_DEBUG" description:"Enable logging AWS API calls"`

	bucket  string
	prefix  string
	session *session.Session
}

func (c *Config) awsConfig() *aws.Config {
	return defaults.Get().Config.
		WithUseDualStack(!c.NoDualStack).
		WithLogLevel(c.awsLogLevel())
}

func (c *Config) awsLogLevel() aws.LogLevelType {
	if c.DebugApi {
		return aws.LogDebug
	}
	return aws.LogOff
}

func (c *Config) s3client() *s3.S3 {
	return s3.New(c.session, c.awsConfig())
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
	ret, err := c.s3client().GetObject(&s3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return err
	}
	defer ret.Body.Close()

	if ret.ContentLength != nil {
		tsize := *ret.ContentLength
		c.logf(7, "%s tsize %d", remoteAddr.String(), tsize)
		xfer.SetSize(tsize)
	}
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
	_, err := s3manager.NewUploaderWithClient(c.s3client()).Upload(&s3manager.UploadInput{
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
}
func (h hook) OnFailure(result tftp.TransferStats, err error) {
	addr := net.UDPAddr{IP: result.RemoteAddr, Port: result.Tid}
	h.logf(4, "ERR %s %s: %s", addr.String(), result.Filename, err.Error())
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
		c.Close();
	}
	return conns[0], nil
}

func parseArgs() (config Config, err error) {
	_, err = flags.Parse(&config)
	if err != nil {
		return
	}

	config.bucket, config.prefix, err = parseS3uri(config.Args.S3uri)
	if err != nil {
		return
	}

	return
}

func main() {
	config, err := parseArgs()
	if err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		} else {
			config.log(2, err)
			os.Exit(1)
		}
	}

	session, err := session.NewSession()
	if err != nil {
		config.log(2, err)
		os.Exit(1)
	}
	config.session = session

	conn, err := getConn()
	if err != nil {
		config.log(2, err)
		os.Exit(1)
	}
	config.logf(5, "Listening on %s", conn.LocalAddr().String())

	server := tftp.NewServer(config.handleRead, config.handleWrite)
	server.SetTimeout(time.Duration(config.Timeout) * time.Millisecond)
	server.SetRetries(config.Retries)
	server.SetHook(config.hooks())

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
