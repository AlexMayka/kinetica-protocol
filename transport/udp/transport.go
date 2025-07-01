package udp

import (
	"context"
	"errors"
	"fmt"
	"kinetica-protocol/transport"
	"net"
)

const network = "udp"

type TransportUDP struct {
	Config   *transport.Config
	listener *net.UDPConn
	ctx      context.Context
	cancel   context.CancelFunc
}

func NewTransportUDP(config *transport.Config) transport.Transport {
	ctx, cancel := context.WithCancel(context.Background())
	return &TransportUDP{Config: config, ctx: ctx, cancel: cancel}
}

func (t *TransportUDP) ConvertAddr() (*net.UDPAddr, error) {
	if t.Config == nil {
		return nil, fmt.Errorf("%w: config is nil", transport.ErrInvalidConfig)
	}
	if t.Config.Address == "" {
		return nil, fmt.Errorf("%w: empty address", transport.ErrInvalidAddrTrans)
	}

	udpAddr, err := net.ResolveUDPAddr(network, t.Config.Address)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to resolve address '%s': %w", transport.ErrInvalidAddrTrans, t.Config.Address, err)
	}

	return udpAddr, nil
}

func (t *TransportUDP) Type() transport.TypeTransport {
	return transport.UDP
}

func (t *TransportUDP) Connection() (transport.Connection, error) {
	udpAddr, err := t.ConvertAddr()
	if err != nil {
		return nil, err
	}

	conn, err := net.DialUDP(network, nil, udpAddr)
	if err != nil {
		var netErr net.Error
		if errors.As(err, &netErr) && netErr.Timeout() {
			return nil, fmt.Errorf("%w: failed to connect to %s: %w", transport.ErrConnectionTimeout, udpAddr.String(), err)
		}

		return nil, fmt.Errorf("%w: failed to connect to %s: %w", transport.ErrConnectionFailed, udpAddr.String(), err)
	}

	return NewConnectionUDP(conn, t.ctx, t.Config.WriteTimeout, t.Config.ReadTimeout), nil
}

func (t *TransportUDP) Listen() (<-chan transport.Connection, error) {
	udpAddr, err := t.ConvertAddr()
	if err != nil {
		return nil, err
	}

	listener, err := net.ListenUDP(network, udpAddr)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to listen on %s: %w", transport.ErrListenAccept, udpAddr.String(), err)
	}

	t.listener = listener

	ch := make(chan transport.Connection, 1)

	udpConn := NewConnectionUDP(listener, t.ctx, t.Config.WriteTimeout, t.Config.ReadTimeout)
	ch <- udpConn

	go func() {
		<-t.ctx.Done()
		close(ch)
	}()

	return ch, nil
}

func (t *TransportUDP) Close() error {
	t.cancel()
	if t.listener != nil {
		return t.listener.Close()
	}

	return nil
}
