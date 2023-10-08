package tronscan

import (
	"context"
	"testing"
	"time"

	"github.com/hardstylez72/bakso_ayam/pkg/log"
	"github.com/stretchr/testify/assert"
)

func TestClient(t *testing.T) {
	c := NewClient()
	c.Logger = &log.Simple{}
	txs, err := c.Txs(context.Background(), time.Now().Add(-time.Minute*20), "TYASr5UV6HEcXatwdFQfmLVUqQQQMUxHLS")
	assert.NoError(t, err)
	assert.NotNil(t, txs)
}
