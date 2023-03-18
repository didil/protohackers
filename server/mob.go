package server

import (
	"bufio"
	"bytes"
	"context"
	"net"
	"os"
	"regexp"
	"strings"
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
		sc.Split(ScanLinesNoLastLine)
		for sc.Scan() {
			msgFromUpstream := sc.Bytes()
			s.logger.Info("message from upstream", zap.ByteString("msgFromUpstream", msgFromUpstream))

			msgToClient := replaceWithBogusCoin2(string(msgFromUpstream))
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
		sc.Split(ScanLinesNoLastLine)

		for sc.Scan() {
			msgFromClient := sc.Bytes()
			s.logger.Info("message from client", zap.ByteString("msgFromClient", msgFromClient))

			msgToUpstream := replaceWithBogusCoin2(string(msgFromClient))
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

var bogusCoinRegex2 = regexp.MustCompile(`^(7[a-zA-Z0-9]{25,34})$`)

func replaceWithBogusCoin2(msg string) string {
	parts := strings.Split(msg, " ")
	modifiedParts := make([]string, 0, len(parts))

	for _, part := range parts {
		if bogusCoinRegex2.MatchString(part) {
			modifiedParts = append(modifiedParts, tonysAddress)
		} else {
			modifiedParts = append(modifiedParts, part)
		}
	}

	return strings.Join(modifiedParts, " ")
}

var bogusCoinRegex = regexp.MustCompile(`(\s|^)(7[a-zA-Z0-9]{25,34})(\s|$)`)

func replaceWithBogusCoin(msg string) string {
	replFunc := func(groups [][]byte) [][]byte {
		groups[1] = []byte(tonysAddress)
		return groups
	}

	return string(replaceAllSubmatchFunc(bogusCoinRegex, []byte(msg), replFunc, -1))
}

// replaceAllSubmatchFunc is the missing regexp.ReplaceAllSubmatchFunc; to use it:
//
//	pattern := regexp.MustCompile(...)
//	data = replaceAllSubmatchFunc(pattern, data, func(groups [][]byte) [][]byte {
//		// mutate groups here
//		return groups
//	})
//
// This snippet is MIT licensed. Please cite by leaving this comment in place. Find
// the latest version at:
//
//	https://gist.github.com/slimsag/14c66b88633bd52b7fa710349e4c6749
func replaceAllSubmatchFunc(re *regexp.Regexp, src []byte, repl func([][]byte) [][]byte, n int) []byte {
	var (
		result  = make([]byte, 0, len(src))
		matches = re.FindAllSubmatchIndex(src, n)
		last    = 0
	)
	for _, match := range matches {
		// Append bytes between our last match and this one (i.e. non-matched bytes).
		matchStart := match[0]
		matchEnd := match[1]
		result = append(result, src[last:matchStart]...)
		last = matchEnd

		// Determine the groups / submatch bytes and indices.
		groups := [][]byte{}
		groupIndices := [][2]int{}
		for i := 2; i < len(match); i += 2 {
			start := match[i]
			end := match[i+1]
			groups = append(groups, src[start:end])
			groupIndices = append(groupIndices, [2]int{start, end})
		}

		// Replace the groups as desired.
		groups = repl(groups)

		// Append match data.
		lastGroup := matchStart
		for i, newValue := range groups {
			// Append bytes between our last group match and this one (i.e. non-group-matched bytes)
			groupStart := groupIndices[i][0]
			groupEnd := groupIndices[i][1]
			result = append(result, src[lastGroup:groupStart]...)
			lastGroup = groupEnd

			// Append the new group value.
			result = append(result, newValue...)
		}
		result = append(result, src[lastGroup:matchEnd]...) // remaining
	}
	result = append(result, src[last:]...) // remaining
	return result
}

func ScanLinesNoLastLine(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF {
		return 0, nil, nil
	}
	if i := bytes.IndexByte(data, '\n'); i >= 0 {
		// We have a full newline-terminated line.
		return i + 1, dropCR(data[0:i]), nil
	}

	// Request more data.
	return 0, nil, nil
}

func dropCR(data []byte) []byte {
	if len(data) > 0 && data[len(data)-1] == '\r' {
		return data[0 : len(data)-1]
	}
	return data
}
