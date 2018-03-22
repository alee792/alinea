package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"time"

	pb "github.com/alee792/alinea/proto"
	"github.com/chromedp/chromedp"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

var (
	port = flag.Int("port", 10000, "The server port")
)

type server struct {
	CDP              *chromedp.CDP
	reloadActive     bool
	reloadCancelChan chan struct{}
	reloadSeconds    int32
}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	s := new()
	pb.RegisterContentPushServer(grpcServer, s)
	log.Infof("Listening on %d.", *port)
	grpcServer.Serve(lis)
}

// New instantiates a server from the given context
func new() *server {
	c, err := chromedp.New(context.Background(), chromedp.WithLog(func(string, ...interface{}) {}))
	if err != nil {
		log.Fatal(err)
	}
	return &server{CDP: c}
}

// PushContent pushes content to a server to display
func (s *server) PushContent(ctx context.Context, c *pb.Content) (*pb.PushResponse, error) {
	s.navigate(c.TargetURL)
	s.hideScrollbar()
	if c.ReloadSeconds > 0 {
		if c.ReloadSeconds > 10 {
			c.ReloadSeconds = 10
		}
		err := s.reloadInterval(time.Duration(c.ReloadSeconds) * time.Second)
		if err != nil {
			return &pb.PushResponse{Success: false}, err
		}
	}

	return &pb.PushResponse{Success: true}, nil
}

// GetContent retrieves the content currently diplayed
func (s *server) GetContent(ctx context.Context, cr *pb.ContentRequest) (*pb.Content, error) {
	c := pb.Content{
		TargetURL:     s.getURL(),
		ReloadSeconds: s.reloadSeconds,
	}
	return &c, nil
}

func (s *server) cancelReload() {
	if s.reloadActive {
		log.Debug("Cancel Reload")
		close(s.reloadCancelChan)
		s.reloadActive = false
	}
}

func (s *server) getURL() (url string) {
	task := chromedp.Tasks{
		chromedp.Location(&url),
	}
	s.CDP.Run(context.Background(), task)
	return
}

func (s *server) hideScrollbar() error {
	var buf []byte
	task := chromedp.Tasks{
		chromedp.Evaluate("document.body.style.overflow = 'hidden'", buf),
	}
	return s.CDP.Run(context.Background(), task)
}

func (s *server) navigate(url string) error {
	s.cancelReload()
	log.WithFields(log.Fields{"target": url}).Infof("Navigating to content.")
	return s.CDP.Run(context.Background(), chromedp.Tasks{chromedp.Navigate(url), chromedp.Sleep(1 * time.Second)})
}

func (s *server) reload() error {
	log.Debug("Reload")
	task := chromedp.Tasks{
		chromedp.Reload(),
	}
	return s.CDP.Run(context.Background(), task)
}

func (s *server) reloadInterval(interval time.Duration) (err error) {
	timeChan := time.Tick(interval)
	s.reloadSeconds = int32(interval.Seconds())
	s.reloadCancelChan = make(chan struct{})
	s.reloadActive = true
	go func() {
		for {
			select {
			case <-s.reloadCancelChan:
				log.Debug("Cancel Reload")
				return
			case <-timeChan:
				s.reload()
			}
		}
	}()
	return
}
