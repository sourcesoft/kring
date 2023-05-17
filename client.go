package kring

import (
	"errors"
	"time"
	// "github.com/mailgun/groupcache"
)

type Options struct {
	// Current nodes name.
	// If not specified defaults to current IP of the pod.
	// If could not detect current IP, it will use hostname
	Name string
	// Open port for broadcasting and advertising.
	// If not specified an available port will be chosen
	Port int
	// Kubernetes headless service config
	KubeHeadlessServiceURL string
	Memberlist             MemberlistConfig
}

type Client struct {
	Options
	Members *Memberlist
}

type OptFunc func(*Options)

func defaultOptions() Options {
	return Options{
		Memberlist: MemberlistConfig{
			GossipNodes:          3,
			GossipInterval:       200 * time.Millisecond, // Gossip more rapidly
			GossipToTheDeadTime:  30 * time.Second,       // Same as push/pull
			GossipVerifyIncoming: true,
			GossipVerifyOutgoing: true,
		},
	}
}

func WithServiceName(serviceName string) OptFunc {
	return func(opts *Options) {
		opts.KubeHeadlessServiceURL = serviceName
	}
}

func validateOptions(opts Options) error {
	if opts.KubeHeadlessServiceURL != "" {
		return errors.New("ServiceName is required")
	}
	return nil
}

func NewClient(opts ...OptFunc) (*Client, error) {
	o := defaultOptions()
	for _, fn := range opts {
		fn(&o)
	}

	err := validateOptions(o)
	if err != nil {
		return nil, err
	}

	members, err := StartGossip(&o)
	if err != nil {
		return nil, err
	}

	c := Client{
		Options: o,
		Members: members,
	}

	return &c, nil
}
