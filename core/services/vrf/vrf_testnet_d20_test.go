package vrf_test

import (
	"math/big"
	"testing"

	"github.com/smartcontractkit/chainlink/core/internal/gethwrappers/generated/vrf_testnet_d20"
	"github.com/smartcontractkit/chainlink/core/store/models"
	"github.com/stretchr/testify/require"
)

func TestD20ExampleContract(t *testing.T) {
	coordinator := deployCoordinator(t)
	keyHash, _, _ := registerProvingKey(t, coordinator)
	d20Address, _, d20, err := vrf_testnet_d20.DeployVRFTestnetD20(coordinator.sergey,
		coordinator.backend, coordinator.rootContractAddress,
		coordinator.linkContractAddress, keyHash)
	require.NoError(t, err, "failed to deploy example contract")
	coordinator.backend.Commit()
	_, err = coordinator.linkContract.Transfer(coordinator.sergey, d20Address,
		big.NewInt(1).Lsh(oneEth, 4))
	require.NoError(t, err, "failed to func example contract with LINK")
	coordinator.backend.Commit()
	// negative control: make sure there are no results prior to rolling
	_, err = d20.D20Results(nil, big.NewInt(0))
	require.Error(t, err, "should be no results in D20 contract")
	_, err = d20.RollDice(coordinator.sergey, seed)
	require.NoError(t, err, "failed to initiate VRF randomness request")
	coordinator.backend.Commit()
	log, err := coordinator.rootContract.FilterRandomnessRequest(nil, nil)
	require.NoError(t, err, "failed to subscribe to RandomnessRequest logs")
	logCount := 0
	for log.Next() {
		logCount += 1
	}
	require.Equal(t, 1, logCount,
		"unexpected log generated by randomness request to VRFCoordinator")
	cLog := models.RawRandomnessRequestLogToRandomnessRequestLog(
		(*models.RawRandomnessRequestLog)(log.Event))
	_ = fulfillRandomnessRequest(t, coordinator, *cLog)
	result, err := d20.D20Results(nil, big.NewInt(0))
	require.NoError(t, err, "failed to retrieve result from D20 contract")
	require.LessOrEqual(t, result.Cmp(big.NewInt(20)), 0)
	require.GreaterOrEqual(t, result.Cmp(big.NewInt(0)), 0)
}
