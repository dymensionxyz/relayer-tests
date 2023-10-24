package interchaintest_test

import (
	"context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"testing"

	relayerinterchaintest "github.com/cosmos/relayer/v2/interchaintest"
	"github.com/cosmos/relayer/v2/relayer"
	interchaintest "github.com/strangelove-ventures/interchaintest/v7"
	"github.com/strangelove-ventures/interchaintest/v7/conformance"
	"github.com/strangelove-ventures/interchaintest/v7/ibc"
	interchaintestrelayer "github.com/strangelove-ventures/interchaintest/v7/relayer"
	"github.com/strangelove-ventures/interchaintest/v7/testreporter"
	"go.uber.org/zap/zaptest"
)

// interchaintestConformance runs the interchaintest conformance tests against
// the provided RelayerFactory.
//
// This is meant to be a relatively quick sanity check,
// so it uses only one pair of chains.
//
// The canonical set of test chains are defined in the interchaintest repository.
func interchaintestConformance(t *testing.T, rf interchaintest.RelayerFactory) {
	numFullNodes := 0
	numValidators := 1
	cf := interchaintest.NewBuiltinChainFactory(zaptest.NewLogger(t), []*interchaintest.ChainSpec{
		{Name: "dymension", Version: "latest", NumFullNodes: &numFullNodes, NumValidators: &numValidators,
			ChainConfig: ibc.ChainConfig{
				Type:           "cosmos",
				ChainID:        "dymension_100-1",
				Images:         []ibc.DockerImage{{Repository: "dymension", UidGid: "1025:1025"}},
				Bin:            "dymd",
				Bech32Prefix:   "dym",
				Denom:          "udym",
				GasPrices:      "0udym",
				GasAdjustment:  0,
				TrustingPeriod: "168h0m0s",
				ModifyGenesisAmounts: func() (sdk.Coin, sdk.Coin) {
					return sdk.NewInt64Coin("udym", 1_000_000_000000000000000000),
						sdk.NewInt64Coin("udym", 500_000_000000000000000000)
				},
			}},
		{Name: "osmosis", Version: "v7.2.0", ChainConfig: ibc.ChainConfig{ChainID: "osmosis-1001"},
			NumFullNodes: &numFullNodes, NumValidators: &numValidators},
	})
	conformance.Test(
		t,
		context.Background(),
		[]interchaintest.ChainFactory{cf},
		[]interchaintest.RelayerFactory{rf},
		testreporter.NewNopReporter(),
	)
}

// TestRelayerInProcess runs the interchaintest conformance tests against
// the current state of this relayer implementation running in process.
func TestRelayerInProcess(t *testing.T) {
	t.Parallel()
	interchaintestConformance(t, relayerinterchaintest.RelayerFactory{})
}

// TestRelayerDockerEventProcessor runs the interchaintest conformance tests against
// the current state of this relayer implementation built in docker.
// Relayer runs using the event processor.
func TestRelayerDockerEventProcessor(t *testing.T) {
	t.Parallel()

	image := relayerinterchaintest.BuildRelayerImage(t)
	rf := interchaintest.NewBuiltinRelayerFactory(
		ibc.CosmosRly,
		zaptest.NewLogger(t),
		interchaintestrelayer.CustomDockerImage(image, "latest", "100:1000"),
		interchaintestrelayer.ImagePull(false),
		interchaintestrelayer.StartupFlags("--processor", "events", "--block-history", "100"),
	)

	interchaintestConformance(t, rf)
}

// TestRelayerDockerLegacyProcessor runs the interchaintest conformance tests against
// the current state of this relayer implementation built in docker.
// Relayer runs using the legacy processor.
func TestRelayerDockerLegacyProcessor(t *testing.T) {
	t.Parallel()
	image := relayerinterchaintest.BuildRelayerImage(t)

	rf := interchaintest.NewBuiltinRelayerFactory(
		ibc.CosmosRly,
		zaptest.NewLogger(t),
		interchaintestrelayer.CustomDockerImage(image, "latest", "100:1000"),
		interchaintestrelayer.ImagePull(false),
		interchaintestrelayer.StartupFlags("--processor", "legacy"),
	)

	interchaintestConformance(t, rf)
}

// TestRelayerEventProcessor runs the interchaintest conformance tests against
// the local relayer code. This is helpful for detecting race conditions.
// Relayer runs using the event processor.
func TestRelayerEventProcessor(t *testing.T) {
	t.Parallel()

	interchaintestConformance(t, relayerinterchaintest.NewRelayerFactory(relayerinterchaintest.RelayerConfig{
		Processor:           relayer.ProcessorEvents,
		InitialBlockHistory: 0,
	}))
}

// TestRelayerLegacyProcessor runs the interchaintest conformance tests against
// the local relayer code. This is helpful for detecting race conditions.
// Relayer runs using the legacy processor.
func TestRelayerLegacyProcessor(t *testing.T) {
	t.Parallel()

	interchaintestConformance(t, relayerinterchaintest.NewRelayerFactory(relayerinterchaintest.RelayerConfig{
		Processor: relayer.ProcessorLegacy,
	}))
}
