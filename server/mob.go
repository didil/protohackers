package server

import (
	"context"
	"net"
	"os"
	"regexp"
	"sync"

	"go.uber.org/zap"
)

const mobMsgLimit = 1024

func (s *Server) HandleMobInTheMiddle(ctx context.Context, conn net.Conn) {
	defer conn.Close()

	var mobUpstreamHost = os.Getenv("MOB_UPSTREAM_HOST")
	var mobUpstreamPort = 16963

	ips, err := net.LookupIP(mobUpstreamHost)
	if err != nil {
		s.logger.Error("chat server dns lookup error", zap.Error(err))
		return
	}
	if len(ips) == 0 {
		s.logger.Error("chat server dns lookup no ips")
		return
	}

	var ip *net.IP

	for _, x := range ips {
		if x.To4() != nil {
			ip = &x
			break
		}
	}

	if ip == nil {
		s.logger.Error("couldn't find ipv4 for chat server")
		return
	}

	upstreamConn, err := net.DialTCP("tcp", nil, &net.TCPAddr{IP: ips[0], Port: mobUpstreamPort})
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
		for {
			buf := make([]byte, mobMsgLimit)
			n, err := upstreamConn.Read(buf)
			if err != nil {
				s.logger.Error("read from upstream error", zap.Error(err))
				break
			}

			_, err = conn.Write(buf[:n])
			if err != nil {
				s.logger.Error("write to client error", zap.Error(err))
				break
			}
		}

		wg.Done()
	}()

	wg.Add(1)
	go func() {
		defer conn.Close()
		defer upstreamConn.Close()
		for {
			buf := make([]byte, mobMsgLimit)
			n, err := conn.Read(buf)
			if err != nil {
				s.logger.Error("read from client error", zap.Error(err))
				return
			}

			_, err = upstreamConn.Write(buf[:n])
			if err != nil {
				s.logger.Error("write to upstream error", zap.Error(err))
				break
			}
		}
		wg.Done()
	}()

	wg.Wait()
}

var bogusCoinRegex = regexp.MustCompile("(7[a-zA-Z0-9]{25,34})")

func replaceBogusCoin(msg string) string {
	bogusCoinRegex.FindAllSubmatchIndex([]byte(msg), -1)
}
