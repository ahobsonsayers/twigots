package twigots

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCountryMarshalJSON(t *testing.T) {
	data, err := json.Marshal(CountryUnitedKingdom)
	require.NoError(t, err)
	require.Equal(t, `"GB"`, string(data))
}

func TestRegionMarshalJSON(t *testing.T) {
	data, err := json.Marshal(RegionLondon)
	require.NoError(t, err)
	require.Equal(t, `"GBLO"`, string(data))
}
