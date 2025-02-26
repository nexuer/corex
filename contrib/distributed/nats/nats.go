package nats

import "github.com/nats-io/nats.go"

type Config struct {
	Endpoints []string
	// TLS credentials
	KeyFile       string
	CertFile      string
	TrustedCAFile string
}

type Nats struct {
	natsConn *nats.Conn
}

func NewNats(config Config) (*Nats, error) {
	nats.Connect("")
	return &Nats{}, nil
}
