package kring

import (
	"errors"
	"fmt"
	"log"
	"net"

	"github.com/hashicorp/memberlist"
)

type MemberlistConfig memberlist.Config
type Memberlist memberlist.Memberlist

func getInitialNodeIPs(serviceName string) ([]string, error) {
	return net.LookupHost(serviceName)
}

type HealthCheck struct {
	Name string
	OK   bool
}

type Delegate struct {
	HealthChecks []HealthCheck
}

func (d *Delegate) NodeMeta(limit int) []byte {
	// Implement this method to broadcast metadata about your node
	return []byte{}
}

func (d *Delegate) NotifyMsg(buf []byte) {
	// Implement this method to handle incoming messages from other nodes
	// For simplicity, let's just print them.
	fmt.Println("Received:", string(buf))
}

func (d *Delegate) GetBroadcasts(overhead, limit int) [][]byte {
	// Implement this method to broadcast your health checks to other nodes
	// For simplicity, we'll send a static message.
	return [][]byte{[]byte("OK")}
}

func (d *Delegate) LocalState(join bool) []byte {
	// Implement this method to share your local state with other nodes
	// This could be anything relevant to your application.
	return []byte("OK")
}

func (d *Delegate) MergeRemoteState(buf []byte, join bool) {
	// Implement this method to handle state updates from other nodes
	// For simplicity, let's just print them.
	fmt.Println("State:", string(buf))
}

func StartGossip(options *Options) (*Memberlist, error) {
	// Prepare delegate.
	delegate := &Delegate{}

	// Create configuration.
	cfg := memberlist.DefaultLocalConfig()
	cfg.Name = options.Name
	cfg.BindPort = options.Port
	cfg.AdvertisePort = cfg.BindPort
	cfg.GossipNodes = options.Memberlist.GossipNodes
	cfg.GossipInterval = options.Memberlist.GossipInterval
	cfg.GossipToTheDeadTime = options.Memberlist.GossipToTheDeadTime
	cfg.GossipVerifyIncoming = options.Memberlist.GossipVerifyIncoming
	cfg.GossipVerifyOutgoing = options.Memberlist.GossipVerifyOutgoing
	cfg.Delegate = delegate

	// Create a new memberlist.
	list, err := memberlist.Create(cfg)
	if err != nil {
		return nil, errors.New("Failed to create memberlist: " + err.Error())
	}

	// Join an existing cluster by specifying at least one known member.
	ips, err := getInitialNodeIPs(options.KubeHeadlessServiceURL)
	if err != nil {
		return nil, errors.New("Failed to fetch IPs of nodes from Kubernetes: " + err.Error())
	}
	n, err := list.Join(ips)
	if err != nil {
		return nil, errors.New("Failed to join cluter: " + err.Error())
	}

	log.Printf("Joined cluster with %d nodes\n", n)

	// Ask for members of the cluster
	for _, member := range list.Members() {
		log.Printf("Member: %s %s\n", member.Name, member.Addr)
	}

	members := Memberlist(*list)

	return &members, nil
}