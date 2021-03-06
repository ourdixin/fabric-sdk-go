/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabpvdr

import (
	"fmt"
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/cryptosuite/bccsp/sw"
	coreMocks "github.com/hyperledger/fabric-sdk-go/pkg/core/mocks"

	"github.com/hyperledger/fabric-sdk-go/pkg/core/cryptosuite"
	fabImpl "github.com/hyperledger/fabric-sdk-go/pkg/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	peerImpl "github.com/hyperledger/fabric-sdk-go/pkg/fab/peer"
	mspImpl "github.com/hyperledger/fabric-sdk-go/pkg/msp"
	mspmocks "github.com/hyperledger/fabric-sdk-go/pkg/msp/test/mockmsp"
	"github.com/stretchr/testify/assert"
)

type mockClientContext struct {
	context.Providers
	msp.SigningIdentity
}

func TestCreateInfraProvider(t *testing.T) {
	newInfraProvider(t)
}

func verifyPeer(t *testing.T, peer fab.Peer, url string) {
	_, ok := peer.(*peerImpl.Peer)
	if !ok {
		t.Fatalf("Unexpected peer impl created")
	}

	// Brittle tests follow
	a := peer.URL()

	if a != url {
		t.Fatalf("Unexpected URL %s", a)
	}
}

func TestCreatePeerFromConfig(t *testing.T) {
	p := newInfraProvider(t)

	url := "grpc://localhost:9999"

	peerCfg := fab.NetworkPeer{
		PeerConfig: fab.PeerConfig{
			URL: url,
		},
	}

	peer, err := p.CreatePeerFromConfig(&peerCfg)

	if err != nil {
		t.Fatalf("Unexpected error creating peer %v", err)
	}

	verifyPeer(t, peer, url)
}

func TestCreateMembership(t *testing.T) {
	p := newInfraProvider(t)
	ctx := mocks.NewMockProviderContext()
	user := mspmocks.NewMockSigningIdentity("user", "user")
	clientCtx := &mockClientContext{
		Providers:       ctx,
		SigningIdentity: user,
	}

	m, err := p.CreateChannelMembership(clientCtx, "test")
	assert.Nil(t, err)
	assert.NotNil(t, m)
}

func TestResolveEventServiceType(t *testing.T) {
	ctx := mocks.NewMockContext(mspmocks.NewMockSigningIdentity("test", "Org1MSP"))
	chConfig := mocks.NewMockChannelCfg("mychannel")

	useDeliver, err := useDeliverEvents(ctx, chConfig)
	assert.NoError(t, err)
	assert.Falsef(t, useDeliver, "expecting deliver events not to be used")

	chConfig.MockCapabilities[fab.ApplicationGroupKey][fab.V1_1Capability] = true

	useDeliver, err = useDeliverEvents(ctx, chConfig)
	assert.NoError(t, err)
	assert.Truef(t, useDeliver, "expecting deliver events to be used")
}

func newInfraProvider(t *testing.T) *InfraProvider {
	configBackend, err := config.FromFile("../../../../test/fixtures/config/config_test.yaml")()
	if err != nil {
		t.Fatalf("config.FromFile failed: %v", err)
	}

	cryptoCfg := cryptosuite.ConfigFromBackend(configBackend...)
	if err != nil {
		t.Fatalf(err.Error())
	}

	endpointCfg, err := fabImpl.ConfigFromBackend(configBackend...)
	if err != nil {
		t.Fatalf(err.Error())
	}

	identityCfg, err := mspImpl.ConfigFromBackend(configBackend...)
	if err != nil {
		t.Fatalf(err.Error())
	}

	cryptoSuite, err := sw.GetSuiteByConfig(cryptoCfg)
	if err != nil {
		panic(fmt.Sprintf("cryptosuiteimpl.GetSuiteByConfig: %v", err))
	}
	im := make(map[string]msp.IdentityManager)
	im[""] = &mocks.MockIdentityManager{}

	ctx := mocks.NewMockProviderContextCustom(cryptoCfg, endpointCfg, identityCfg, cryptoSuite, coreMocks.NewMockSigningManager(), &mspmocks.MockUserStore{}, im)
	ip := New(endpointCfg)
	ip.Initialize(ctx)

	return ip
}
