package server

import (
	"bufio"
	"context"
	"net"
	"os"
	"regexp"
	"sync"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

const mobMsgLimit = 1024

func (s *Server) HandleMobInTheMiddle(ctx context.Context, conn net.Conn) {
	defer conn.Close()

	tcpAddr, err := getMobUpstreamTcpAddr()
	if err != nil {
		s.logger.Error("getMobUpstreamTcpAddr error", zap.Error(err))
		return
	}

	upstreamConn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		s.logger.Error("chat server connect error", zap.Error(err))
		return
	}
	defer upstreamConn.Close()

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer conn.Close()
		defer upstreamConn.Close()
		sc := bufio.NewScanner(upstreamConn)
		for sc.Scan() {
			msgFromUpstream := sc.Bytes()
			s.logger.Info("message from upstream", zap.ByteString("msgFromUpstream", msgFromUpstream))

			msgToClient := replaceWithBogusCoin(string(msgFromUpstream)) + "\n"
			s.logger.Info("message to client", zap.String("msgToClient", msgToClient))

			_, err = conn.Write([]byte(msgToClient + "\n"))
			if err != nil {
				s.logger.Error("write to client error", zap.Error(err))
				break
			}
		}

		err = sc.Err()
		if err != nil {
			s.logger.Error("read from upstream error", zap.Error(err))
		}

		wg.Done()
	}()

	wg.Add(1)
	go func() {
		defer conn.Close()
		defer upstreamConn.Close()
		sc := bufio.NewScanner(conn)
		for sc.Scan() {
			msgFromClient := sc.Bytes()
			s.logger.Info("message from client", zap.ByteString("msgFromClient", msgFromClient))

			msgToUpstream := replaceWithBogusCoin(string(msgFromClient))
			s.logger.Info("message to upstream", zap.String("msgToUpstream", msgToUpstream))

			_, err = upstreamConn.Write([]byte(msgToUpstream + "\n"))
			if err != nil {
				s.logger.Error("write to upstream error", zap.Error(err))
				break
			}
		}

		err = sc.Err()
		if err != nil {
			s.logger.Error("read from client error", zap.Error(err))
		}

		wg.Done()
	}()

	wg.Wait()
}

var mobUpstreamPort = 16963

func getMobUpstreamTcpAddr() (*net.TCPAddr, error) {
	var mobUpstreamHost = os.Getenv("MOB_UPSTREAM_HOST")

	ips, err := net.LookupIP(mobUpstreamHost)
	if err != nil {
		return nil, errors.Wrapf(err, "chat server dns lookup error")
	}
	if len(ips) == 0 {
		return nil, errors.Wrapf(err, "chat server dns lookup no ips")
	}

	var ip *net.IP

	for _, x := range ips {
		if x.To4() != nil {
			ip = &x
			break
		}
	}

	if ip == nil {
		return nil, errors.New("couldn't find ipv4 for chat server")
	}

	return &net.TCPAddr{IP: ips[0], Port: mobUpstreamPort}, nil
}

var tonysAddress = "7YWHMfk9JZe0LM0g1ZauHuiSxhI"

var bogusCoinRegex = regexp.MustCompile("\\b(7[a-zA-Z0-9]{25,34})\\b")

func replaceWithBogusCoin(msg string) string {
	return bogusCoinRegex.ReplaceAllString(msg, tonysAddress)
}
