package grpc

import (
	"context"
	"github.com/JMURv/sso/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
	"sync"
	"testing"
	"time"
)

func TestStartAndClose(t *testing.T) {
	mock := gomock.NewController(t)
	defer mock.Finish()

	mctrl := mocks.NewMockCtrl(mock)
	auth := mocks.NewMockAuthService(mock)

	h := New(auth, mctrl)

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		h.Start(50051)
	}()

	time.Sleep(100 * time.Millisecond)

	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	require.NoError(t, err)

	client := grpc_health_v1.NewHealthClient(conn)
	resp, err := client.Check(
		context.Background(), &grpc_health_v1.HealthCheckRequest{
			Service: "sso",
		},
	)
	require.NoError(t, err)
	assert.Equal(t, grpc_health_v1.HealthCheckResponse_SERVING, resp.Status)

	conn.Close()
	h.Close()

	wg.Wait()
}
