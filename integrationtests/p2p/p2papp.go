package p2p

import (
	"math/rand"
	"time"

	"github.com/libp2p/go-libp2p-peerstore"
	"github.com/phoreproject/synapse/integrationtests/framework"
	"github.com/phoreproject/synapse/p2p/app"
)

var startPort = 19000
var appCount = 10
var peerCount = 1

// TestCase implements IntegrationTest
type TestCase struct {
	appList []*app.App
}

// Execute implements IntegrationTest
func (test TestCase) Execute(service *testframework.TestService) error {
	for i := 0; i < appCount; i++ {
		app := test.createApp(i)
		test.appList = append(test.appList, app)
	}

	for _, app := range test.appList {
		go app.Run()
		time.Sleep(100 * time.Millisecond)
	}

	// Wait until host nodes are created
	time.Sleep(300 * time.Millisecond)

	test.connectApps()

	for {
		time.Sleep(100 * time.Millisecond)
	}
}

func (test TestCase) createApp(index int) *app.App {
	config := app.NewConfig()
	config.ListeningPort = startPort + index
	config.MinPeerCountToWait = 0
	config.HeartBeatInterval = 10 * 1000
	app := app.NewApp(config)

	return app
}

func (test TestCase) connectApps() {
	for i := 0; i < len(test.appList); i++ {
		for k := 0; k < peerCount; k++ {
			peerIndex := rand.Int() % len(test.appList)
			test.appList[i].GetHostNode().Connect(&peerstore.PeerInfo{
				ID:    test.appList[peerIndex].GetHostNode().GetHost().ID(),
				Addrs: test.appList[peerIndex].GetHostNode().GetHost().Addrs(),
			})
		}
	}
}
