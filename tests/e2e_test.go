package tests

import (
	"context"
	"fmt"
	"net"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/etcd"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/davidseybold/beacondns/client"
)

const (
	postgresDBName     = "beacondns"
	postgresDBUser     = "beacondns"
	postgresDBPassword = "beacondns"

	controllerPort = "8080"
)

func TestE2E(t *testing.T) {
	ctx := t.Context()
	n, err := network.New(ctx)
	if err != nil {
		t.Fatalf("failed to create network: %v", err)
	}
	defer n.Remove(ctx)

	postgresContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase(postgresDBName),
		postgres.WithUsername(postgresDBUser),
		postgres.WithPassword(postgresDBPassword),
		postgres.BasicWaitStrategies(),
		network.WithNetwork([]string{"postgres-db"}, n),
	)
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}
	defer func() {
		if terminateErr := testcontainers.TerminateContainer(postgresContainer); terminateErr != nil {
			t.Logf("failed to terminate postgres container: %v", terminateErr)
		}
	}()

	etcdContainer, err := etcd.Run(ctx, "gcr.io/etcd-development/etcd:v3.5.14",
		network.WithNetwork([]string{"etcd"}, n),
	)
	if err != nil {
		t.Fatalf("failed to start etcd container: %v", err)
	}
	defer func() {
		if terminateErr := testcontainers.TerminateContainer(etcdContainer); terminateErr != nil {
			t.Logf("failed to terminate etcd container: %v", terminateErr)
		}
	}()

	postgresURL := "postgres-db"
	etcdEndpoint := "http://etcd:2379"

	controllerContainer, err := runController(ctx, controllerEnv{
		PostgresURL: postgresURL,
		EtcdURL:     etcdEndpoint,
	}, n)
	if err != nil {
		t.Fatalf("failed to start controller container: %v", err)
	}
	defer func() {
		if terminateErr := testcontainers.TerminateContainer(controllerContainer); terminateErr != nil {
			t.Logf("failed to terminate controller container: %v", terminateErr)
		}
	}()

	resolverContainer, err := runResolver(ctx, resolverEnv{
		EtcdURL: etcdEndpoint,
	}, n)
	if err != nil {
		t.Fatalf("failed to start resolver container: %v", err)
	}
	defer func() {
		if terminateErr := testcontainers.TerminateContainer(resolverContainer); terminateErr != nil {
			t.Logf("failed to terminate resolver container: %v", terminateErr)
		}
	}()

	controllerHost, err := controllerContainer.Host(ctx)
	if err != nil {
		t.Fatalf("failed to get controller host: %v", err)
	}

	controllerPort, err := controllerContainer.MappedPort(ctx, "8080/tcp")
	if err != nil {
		t.Fatalf("failed to get controller port: %v", err)
	}
	controllerPortStr := controllerPort.Port()

	beaconHost := fmt.Sprintf("http://%s", net.JoinHostPort(controllerHost, controllerPortStr))

	beaconClient := client.New(beaconHost)

	t.Logf("beaconHost: %s", beaconHost)

	_, err = beaconClient.GetZone(ctx, "test.com")
	var nze *client.NoSuchZoneError
	require.ErrorAs(t, err, &nze)
	t.Logf("%s\n", err.Error())
}

type controllerEnv struct {
	PostgresURL string
	EtcdURL     string
}

func runController(
	ctx context.Context,
	env controllerEnv,
	network *testcontainers.DockerNetwork,
) (testcontainers.Container, error) {
	projectRoot, err := getProjectRoot()
	if err != nil {
		return nil, err
	}
	controllerContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			FromDockerfile: testcontainers.FromDockerfile{
				Context:    projectRoot,
				Dockerfile: "tests/docker/Dockerfile.controller",
				KeepImage:  true,
			},
			Env: map[string]string{
				"BEACON_DB_HOST":         env.PostgresURL,
				"BEACON_DB_NAME":         postgresDBName,
				"BEACON_DB_USER":         postgresDBUser,
				"BEACON_DB_PASSWORD":     postgresDBPassword,
				"BEACON_ETCD_ENDPOINTS":  env.EtcdURL,
				"BEACON_CONTROLLER_PORT": controllerPort,
			},
			ExposedPorts: []string{"8080/tcp"},
			WaitingFor:   wait.ForHTTP("/health").WithPort(controllerPort),
			Networks:     []string{network.Name},
			NetworkAliases: map[string][]string{
				network.Name: {"beacondns-controller"},
			},
			LogConsumerCfg: &testcontainers.LogConsumerConfig{
				Consumers: []testcontainers.LogConsumer{&testcontainers.StdoutLogConsumer{}},
			},
		},
		Started: true,
	})
	if err != nil {
		return nil, err
	}

	return controllerContainer, nil
}

type resolverEnv struct {
	EtcdURL string
}

func runResolver(
	ctx context.Context,
	env resolverEnv,
	network *testcontainers.DockerNetwork,
) (testcontainers.Container, error) {
	projectRoot, err := getProjectRoot()
	if err != nil {
		return nil, err
	}
	resolverContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			FromDockerfile: testcontainers.FromDockerfile{
				Context:    projectRoot,
				Dockerfile: "tests/docker/Dockerfile.resolver",
				KeepImage:  true,
			},
			Env: map[string]string{
				"BEACON_ETCD_ENDPOINTS": env.EtcdURL,
				"BEACON_FORWARDER":      "1.1.1.1",
			},
			ExposedPorts: []string{"53/udp"},
			WaitingFor:   wait.ForLog("starting beacon"),
			Networks:     []string{network.Name},
			NetworkAliases: map[string][]string{
				network.Name: {"beacondns-resolver"},
			},
			LogConsumerCfg: &testcontainers.LogConsumerConfig{
				Consumers: []testcontainers.LogConsumer{&testcontainers.StdoutLogConsumer{}},
			},
		},
		Started: true,
	})
	if err != nil {
		return nil, err
	}

	return resolverContainer, nil
}

func getProjectRoot() (string, error) {
	root, err := filepath.Abs("..")
	if err != nil {
		return "", fmt.Errorf("failed to get project root: %w", err)
	}
	return root, nil
}
