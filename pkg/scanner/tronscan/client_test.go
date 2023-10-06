package tronscan

import (
	"bakso_ayam/pkg/log"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestClient(t *testing.T) {
	c := NewClient()
	c.Logger = &log.Simple{}
	txs, err := c.Txs(context.Background(), time.Now().Add(-time.Minute*20), "TYASr5UV6HEcXatwdFQfmLVUqQQQMUxHLS")
	assert.NoError(t, err)
	assert.NotNil(t, txs)
}
