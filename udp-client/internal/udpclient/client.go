package udpclient

import (
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
)

type Config struct {
	BindStart int
	BindEnd   int
}

func Run(args []string, stdin io.Reader, config Config) error {
	if len(args) < 2 {
		return errors.New("expected ip port")
	}

	port, err := strconv.Atoi(args[1])
	if err != nil {
		return fmt.Errorf("expected numerical port: %w", err)
	}

	payload, err := ReadPayload(args, stdin)
	if err != nil {
		return err
	}

	conn, err := BindClientSocket(config)
	if err != nil {
		return err
	}
	defer conn.Close()

	remoteAddr := &net.UDPAddr{IP: net.ParseIP(args[0]), Port: port}
	if remoteAddr.IP == nil {
		return fmt.Errorf("invalid ip %q", args[0])
	}
	if _, err := conn.WriteToUDP(payload, remoteAddr); err != nil {
		return fmt.Errorf("send udp payload: %w", err)
	}

	return nil
}

func ReadPayload(args []string, stdin io.Reader) ([]byte, error) {
	if len(args) >= 3 {
		return []byte(args[2]), nil
	}

	payload, err := io.ReadAll(stdin)
	if err != nil {
		return nil, fmt.Errorf("read stdin: %w", err)
	}
	if len(payload) == 0 {
		return nil, errors.New("received no data")
	}

	return payload, nil
}

func BindClientSocket(config Config) (*net.UDPConn, error) {
	for port := config.BindStart; port <= config.BindEnd; port++ {
		conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4zero, Port: port})
		if err == nil {
			return conn, nil
		}
	}

	return nil, errors.New("cannot create udp socket")
}
