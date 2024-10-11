package kvs_test

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-kvs-client/kvs"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/iskaypetcom/digital/sre/tools/dev/backend-api-sdk/v2/core"
	"gitlab.com/iskaypetcom/digital/sre/tools/dev/backend-api-sdk/v2/core/application"
	"gitlab.com/iskaypetcom/digital/sre/tools/dev/backend-api-sdk/v2/core/routing"
	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-kvs-client/kvs/dynamodb"
	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-logger/log"
	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/rest"
)

type TestApp struct {
	application.APIApplication
}

func (r *TestApp) Init() {
	r.UseMetrics()
	r.RegisterRoutes(new(Routes))
}

type Routes struct {
	routing.APIRoutes
}

func (r *Routes) Register() {
	lowLevelClient := kvs.NewLowLevelClientProxy(dynamodb.
		NewLowLevelClient(dynamodb.
			NewAWSFakeClient(),
			"users-cache"),
	)

	ctx := context.Background()
	err := lowLevelClient.SaveWithContext(ctx, "my-key", &kvs.Item{
		Key:   "my-key",
		Value: "my-value",
	})
	if err != nil {
		log.Fatal(err)
	}

	r.AddRoute(http.MethodGet, "/kvs/get", func(httpCtx *routing.HTTPContext) error {
		value, kvsErr := lowLevelClient.Get("my-key")
		if kvsErr != nil {
			httpCtx.Status(http.StatusInternalServerError)
			return httpCtx.SendString(kvsErr.Error())
		}

		log.Debugf("Got value: %s", value.Value)

		_, kvsErr = lowLevelClient.Get("missing-key")
		if kvsErr != nil && !errors.Is(kvsErr, kvs.ErrKeyNotFound) {
			httpCtx.Status(http.StatusInternalServerError)
			return httpCtx.SendString(kvsErr.Error())
		}

		return httpCtx.SendString("pong")
	})
}

func TestCollector_IncrementCounter(t *testing.T) {
	t.Setenv("APP_NAME", "kvs-client")
	t.Setenv("ENV", "local")

	listener, receiveErr := net.Listen("tcp", ":0")
	require.NoError(t, receiveErr)

	addr, ok := listener.Addr().(*net.TCPAddr)
	assert.True(t, ok)

	port := addr.Port
	receiveErr = listener.Close()
	require.NoError(t, receiveErr)

	server := core.NewServer(port)
	server.On(new(TestApp))
	server.Start()

	go func() {
		time.Sleep(500 * time.Millisecond)

		requestBuilder := rest.RequestBuilder{
			BaseURL: fmt.Sprintf("http://0.0.0.0:%d", port),
		}

		response := requestBuilder.Get("/kvs/get")
		assert.NoError(t, response.Err)
		assert.Equal(t, http.StatusOK, response.StatusCode)

		response = requestBuilder.Get("/metrics")
		assert.NoError(t, response.Err)
		assert.Equal(t, http.StatusOK, response.StatusCode)

		got := response.String()

		want := fmt.Sprintf(`__kvs_stats{client_name="users-cache",stats="hit"} %d`, 1)
		if !strings.Contains(got, want) {
			t.Errorf("got %s; want %s", got, want)
		}

		want = fmt.Sprintf(`__kvs_stats{client_name="users-cache",stats="miss"} %d`, 1)
		if !strings.Contains(got, want) {
			t.Errorf("got %s; want %s", got, want)
		}

		want = fmt.Sprintf(`__kvs_stats{client_name="users-cache",stats="save"} %d`, 1)
		if !strings.Contains(got, want) {
			t.Errorf("got %s; want %s", got, want)
		}

		want = fmt.Sprintf(`__kvs_stats{client_name="users-cache",stats="save"} %d`, 1)
		if !strings.Contains(got, want) {
			t.Errorf("got %s; want %s", got, want)
		}

		want = fmt.Sprintf(`__kvs_connection_count{client_name="users-cache",type="save"} %d`, 1)
		if !strings.Contains(got, want) {
			t.Errorf("got %s; want %s", got, want)
		}

		assert.NoError(t, server.Shutdown())
	}()

	require.NoError(t, server.Join())
}
