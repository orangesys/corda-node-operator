package utils

import (
	"strconv"
	"strings"

	"github.com/go-akka/configuration"
	cordav1 "orangesys.io/cordanode/api/v1"
)

//NodeInfoParser ...
type NodeInfoParser struct {
	conf *configuration.Config
}

//NewNodeInfoParser ...
func NewNodeInfoParser(cr *cordav1.CordaNode) *NodeInfoParser {
	return &NodeInfoParser{
		conf: configuration.ParseString(string(cr.Spec.NodeInfo)),
	}
}

//GetP2PAddressPort ...
func (n *NodeInfoParser) GetP2PAddressPort() (int32, error) {
	portStr := strings.Split(n.conf.GetString("p2pAddress"), ":")[1]
	i, err := strconv.ParseInt(portStr, 10, 32)
	if err != nil {
		return 0, err
	}
	return int32(i), nil
}
