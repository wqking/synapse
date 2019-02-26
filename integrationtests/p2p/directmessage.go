package p2p

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/phoreproject/synapse/pb"

	crypto "github.com/libp2p/go-libp2p-crypto"
	peerstore "github.com/libp2p/go-libp2p-peerstore"
	ma "github.com/multiformats/go-multiaddr"

	logger "github.com/sirupsen/logrus"

	"github.com/phoreproject/synapse/integrationtests/framework"
	"github.com/phoreproject/synapse/p2p"
)

// DirectMessageTest implements IntegrationTest
type DirectMessageTest struct {
}

type testNodeDirectMessage struct {
	*p2p.HostNode
	nodeID int
}

func keyToString(key crypto.Key) string {
	data, _ := key.Bytes()
	return hex.EncodeToString(data)
}

// Execute implements IntegrationTest
func (test DirectMessageTest) Execute(service *testframework.TestService) error {
	logger.SetLevel(logger.TraceLevel)

	hostNode0, err := createHostNodeForDirectMessage(0)
	if err != nil {
		logger.Warn(err)
		return err
	}
	hostNode1, err := createHostNodeForDirectMessage(1)
	if err != nil {
		logger.Warn(err)
		return err
	}

	connectToPeerForDirectMessage(hostNode0, hostNode1)
	connectToPeerForDirectMessage(hostNode1, hostNode0)

	for i := 0; i < 5; i++ {
		message := fmt.Sprintf("Test message of %d", i)
		fmt.Printf("Request: %s\n", message)
		hostNode0.GetLivePeerList()[0].SendMessage(&pb.TestMessage{Message: "Node0 " + message})
		hostNode1.GetLivePeerList()[0].SendMessage(&pb.TestMessage{Message: "Node1 " + message})

		time.Sleep(500 * time.Millisecond)
	}

	return nil
}

func createNodeAddressForDirectMessage(index int) ma.Multiaddr {
	addr, err := ma.NewMultiaddr(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", 19000+index))
	if err != nil {
		logger.WithField("Function", "createNodeAddressForDirectMessage").Warn(err)
		return nil
	}
	return addr
}

func createHostNodeForDirectMessage(index int) (*testNodeDirectMessage, error) {
	privateKey, publicKey, err := crypto.GenerateSecp256k1Key(rand.Reader)
	if err != nil {
		logger.WithField("Function", "createHostNodeForDirectMessage").Warn(err)
		return nil, err
	}

	hostNode, err := p2p.NewHostNode(createNodeAddressForDirectMessage(index), publicKey, privateKey)
	if err != nil {
		logger.WithField("Function", "createHostNodeForDirectMessage").Warn(err)
		return nil, err
	}

	node := &testNodeDirectMessage{
		HostNode: hostNode,
		nodeID:   index,
	}

	node.RegisterMessageHandler("pb.TestMessage", func(peer *p2p.PeerNode, message proto.Message) {
		logger.Debugf("Node %s received message %s ", keyToString(*node.GetPublicKey()), proto.MessageName(message))
	})
	node.SetOnPeerConnectedHandler(func(peer *p2p.PeerNode) {
		logger.Debugf("Node %s has new connection ", keyToString(*node.GetPublicKey()))
	})

	return node, nil
}

func connectToPeerForDirectMessage(hostNode *testNodeDirectMessage, target *testNodeDirectMessage) *p2p.PeerNode {
	addrs := target.GetHost().Addrs()

	peerInfo := peerstore.PeerInfo{
		ID:    target.GetHost().ID(),
		Addrs: addrs,
	}

	logger.WithField("Function", "connectToPeerForDirectMessage").Trace(peerInfo.ID.Pretty())

	node, err := hostNode.Connect(&peerInfo)
	if err != nil {
		logger.WithField("Function", "connectToPeerForDirectMessage").Warn(err)
		return nil
	}
	return node
}
