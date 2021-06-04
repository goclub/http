package xhttp

import (
	"context"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
	"time"
)

func TestClient_SendRetry(t *testing.T) {
	ctx := context.Background()
	client := NewClient(&http.Client{})
	resp, bodyClose, statusCode, err := client.Send(ctx, GET, "https://bing.com/404", SendRequest{
		Retry: RequestRetry{
			Times: 3,
			Interval:  time.Millisecond*100,
		},
		Debug: true,
	});  ; assert.NoError(t, err)
	defer assert.NoError(t, bodyClose())
	assert.Equal(t, statusCode, 404)
	_=resp
}
