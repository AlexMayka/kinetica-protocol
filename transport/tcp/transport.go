package tcp

import (
	"context"
	"errors"
	"fmt"
	"kinetica-protocol/transport"
	"net"
)

const network = "tcp"

type TransportTCP struct {
	Config   *transport.Config
	listener *net.TCPListener
	ctx      context.Context
	cancel   context.CancelFunc
}

func NewTransportTCP(config *transport.Config) *TransportTCP {
	ctx, cancel := context.WithCancel(context.Background())
	return &TransportTCP{Config: config, ctx: ctx, cancel: cancel}
}

func (t *TransportTCP) ConvertAddr() (*net.TCPAddr, error) {
	if t.Config == nil {
		return nil, fmt.Errorf("%w: config is nil", transport.ErrInvalidConfig)
	}
	if t.Config.Address == "" {
		return nil, fmt.Errorf("%w: empty address", transport.ErrInvalidAddrTrans)
	}

	tcpAddr, err := net.ResolveTCPAddr(network, t.Config.Address)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to resolve address '%s': %w", transport.ErrInvalidAddrTrans, t.Config.Address, err)
	}

	return tcpAddr, nil
}

func (t *TransportTCP) Type() transport.TypeTransport {
	return transport.TCP
}

func (t *TransportTCP) Connection() (transport.Connection, error) {
	tcpAddr, err := t.ConvertAddr()
	if err != nil {
		return nil, err
	}

	conn, err := net.DialTCP(network, nil, tcpAddr)
	if err != nil {
		var netErr net.Error
		if errors.As(err, &netErr) && netErr.Timeout() {
			return nil, fmt.Errorf("%w: failed to connect to %s: %w", transport.ErrConnectionTimeout, tcpAddr.String(), err)
		}
		return nil, fmt.Errorf("%w: failed to connect to %s: %w", transport.ErrConnectionFailed, tcpAddr.String(), err)
	}
	return NewConnectionTCP(conn, t.ctx, t.Config.WriteTimeout, t.Config.ReadTimeout), nil
}

func (t *TransportTCP) Listen() (<-chan transport.Connection, error) {
	tcpAddr, err := t.ConvertAddr()
	if err != nil {
		return nil, err
	}

	listener, err := net.ListenTCP(network, tcpAddr)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to listen on %s: %w", transport.ErrListenAccept, tcpAddr.String(), err)
	}

	t.listener = listener

	ch := make(chan transport.Connection)
	go func() {
		defer close(ch)
		for {
			conn, err := listener.Accept()
			if errors.Is(err, net.ErrClosed) {
				return
			}
			if err != nil {
				continue
			}
			ch <- NewConnectionTCP(conn, t.ctx, t.Config.WriteTimeout, t.Config.ReadTimeout)
		}
	}()

	return ch, nil
}

func (t *TransportTCP) Close() error {
	t.cancel()
	if t.listener != nil {
		return t.listener.Close()
	}

	return nil
}
