package strelets

import (
	"fmt"
	"github.com/orbs-network/boyarin/test"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_getNetworkConfigJSON(t *testing.T) {
	peers := make(PeersMap)

	for i, key := range test.PublicKeys() {
		peers[PublicKey(key)] = &Peer{
			IP:   fmt.Sprintf("10.0.0.%d", i+1),
			Port: 4400 + i,
		}
	}

	require.JSONEq(t, `{
		"federation-nodes": [
			{"Key":"dfc06c5be24a67adee80b35ab4f147bb1a35c55ff85eda69f40ef827bddec173","IP":"10.0.0.1","Port":4400},
			{"Key":"92d469d7c004cc0b24a192d9457836bf38effa27536627ef60718b00b0f33152","IP":"10.0.0.2","Port":4401},
			{"Key":"a899b318e65915aa2de02841eeb72fe51fddad96014b73800ca788a547f8cce0","IP":"10.0.0.3","Port":4402}
		]
	}`, string(getNetworkConfigJSON(&peers)))
}
