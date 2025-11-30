package twigots

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCurrencyMarshalJSON(t *testing.T) {
	data, err := json.Marshal(CurrencyGBP)
	require.NoError(t, err)
	require.Equal(t, `"GBP"`, string(data))
}
