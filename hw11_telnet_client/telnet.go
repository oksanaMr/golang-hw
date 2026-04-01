package main

import (
	"bufio"
	"io"
	"net"
	"time"
)

type TelnetClient interface {
	Connect() error
	io.Closer
	Send() error
	Receive() error
}

type TelnetClientImpl struct {
	address string
	timeout time.Duration
	in      io.ReadCloser
	out     io.Writer
	conn    net.Conn
}

func NewTelnetClient(address string, timeout time.Duration, in io.ReadCloser, out io.Writer) TelnetClient {
	return &TelnetClientImpl{
		address: address,
		timeout: timeout,
		in:      in,
		out:     out,
	}
}

func (t *TelnetClientImpl) Connect() error {
	conn, err := net.DialTimeout("tcp", t.address, t.timeout)
	if err != nil {
		return err
	}
	t.conn = conn
	return nil
}

func (t *TelnetClientImpl) Send() error {
	scanner := bufio.NewScanner(t.in)
	for scanner.Scan() {
		clientMessage := append(scanner.Bytes(), '\n')
		_, err := t.conn.Write(clientMessage)
		if err != nil {
			return err
		}
	}

	return scanner.Err()
}

func (t *TelnetClientImpl) Receive() error {
	scanner := bufio.NewScanner(t.conn)
	for scanner.Scan() {
		clientMessage := append(scanner.Bytes(), '\n')
		_, err := t.out.Write(clientMessage)
		if err != nil {
			return err
		}
	}

	return scanner.Err()
}

func (t *TelnetClientImpl) Close() error {
	if t.conn != nil {
		return t.conn.Close()
	}
	return nil
}
