package kvs_test

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-kvs-client/kvs"
	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-kvs-client/kvs/dynamodb"
	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/rest"
)

var times uint64

var kvsClientProxy = kvs.NewLowLevelClientProxy(dynamodb.
	NewLowLevelClient(dynamodb.
		NewAWSFakeClient(),
		"users-cache"),
)

func TestCollector_IncrementCounter(t *testing.T) {
	addr, err := rndAddr()
	require.NoError(t, err)

	atomic.AddUint64(&times, 1)

	err = kvsClientProxy.SaveWithContext(t.Context(), "my-key", &kvs.Item{
		Key:   "my-key",
		Value: "my-value",
	})

	require.NoError(t, err)

	router := mux.NewRouter()
	router.Handle("/metrics", promhttp.Handler())
	router.HandleFunc("/kvs/get", func(w http.ResponseWriter, r *http.Request) {
		_, err = kvsClientProxy.Get("my-key")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		_, err = kvsClientProxy.Get("missing-key")
		if err != nil && !errors.Is(err, kvs.ErrKeyNotFound) {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	})

	client := rest.Client{
		BaseURL: addr.HTTP,
	}

	server := &http.Server{
		Addr:    addr.HTTP,
		Handler: router,
	}

	go func() {
		time.Sleep(500 * time.Millisecond)

		response := client.GetWithContext(t.Context(), "/kvs/get")
		assert.NoError(t, response.Err)
		assert.Equal(t, http.StatusOK, response.StatusCode)

		response = client.GetWithContext(t.Context(), "/metrics")
		assert.NoError(t, response.Err)
		assert.Equal(t, http.StatusOK, response.StatusCode)

		got := response.String()

		want := fmt.Sprintf(`__kvs_stats{client_name="users-cache",stats="hit"} %d`, 1*times)
		if !strings.Contains(got, want) {
			t.Errorf("got %s; want %s", got, want)
		}

		want = fmt.Sprintf(`__kvs_stats{client_name="users-cache",stats="miss"} %d`, 1*times)
		if !strings.Contains(got, want) {
			t.Errorf("got %s; want %s", got, want)
		}

		want = fmt.Sprintf(`__kvs_stats{client_name="users-cache",stats="save"} %d`, 1*times)
		if !strings.Contains(got, want) {
			t.Errorf("got %s; want %s", got, want)
		}

		want = fmt.Sprintf(`__kvs_stats{client_name="users-cache",stats="save"} %d`, 1*times)
		if !strings.Contains(got, want) {
			t.Errorf("got %s; want %s", got, want)
		}

		want = fmt.Sprintf(`__kvs_connection_count{client_name="users-cache",type="save"} %d`, 1*times)
		if !strings.Contains(got, want) {
			t.Errorf("got %s; want %s", got, want)
		}

		assert.NoError(t, server.Shutdown(t.Context()))
	}()

	err = server.Serve(addr.Listener)
	require.Error(t, err)
	require.ErrorIs(t, err, http.ErrServerClosed)
}

type Addr struct {
	Listener net.Listener
	Host     string
	HTTP     string
	Port     int
}

func rndAddr() (*Addr, error) {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return nil, err
	}

	addr, ok := listener.Addr().(*net.TCPAddr)
	if !ok {
		return nil, fmt.Errorf("invalid address type: %T", listener.Addr())
	}

	dfltHost := "0.0.0.0"

	return &Addr{
		Listener: listener,
		Host:     dfltHost,
		Port:     addr.Port,
		HTTP:     fmt.Sprintf("http://%s", net.JoinHostPort(dfltHost, strconv.Itoa(addr.Port))),
	}, nil
}
