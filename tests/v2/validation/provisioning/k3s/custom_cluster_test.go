package k3s

import (
	"testing"

	"github.com/rancher/rancher/tests/framework/clients/rancher"
	management "github.com/rancher/rancher/tests/framework/clients/rancher/generated/management/v3"
	"github.com/rancher/rancher/tests/framework/extensions/clusters"
	"github.com/rancher/rancher/tests/framework/extensions/clusters/kubernetesversions"
	"github.com/rancher/rancher/tests/framework/extensions/machinepools"
	"github.com/rancher/rancher/tests/framework/extensions/provisioninginput"
	"github.com/rancher/rancher/tests/framework/extensions/users"
	password "github.com/rancher/rancher/tests/framework/extensions/users/passwordgenerator"
	"github.com/rancher/rancher/tests/framework/pkg/config"
	namegen "github.com/rancher/rancher/tests/framework/pkg/namegenerator"
	"github.com/rancher/rancher/tests/framework/pkg/session"
	"github.com/rancher/rancher/tests/v2/validation/provisioning/permutations"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type CustomClusterProvisioningTestSuite struct {
	suite.Suite
	client             *rancher.Client
	session            *session.Session
	standardUserClient *rancher.Client
	clustersConfig     *provisioninginput.Config
}

func (c *CustomClusterProvisioningTestSuite) TearDownSuite() {
	c.session.Cleanup()
}

func (c *CustomClusterProvisioningTestSuite) SetupSuite() {
	testSession := session.NewSession()
	c.session = testSession

	c.clustersConfig = new(provisioninginput.Config)
	config.LoadConfig(provisioninginput.ConfigurationFileKey, c.clustersConfig)

	client, err := rancher.NewClient("", testSession)
	require.NoError(c.T(), err)

	c.client = client

	c.clustersConfig.K3SKubernetesVersions, err = kubernetesversions.Default(c.client, clusters.K3SClusterType.String(), c.clustersConfig.K3SKubernetesVersions)
	require.NoError(c.T(), err)

	enabled := true
	var testuser = namegen.AppendRandomString("testuser-")
	var testpassword = password.GenerateUserPassword("testpass-")
	user := &management.User{
		Username: testuser,
		Password: testpassword,
		Name:     testuser,
		Enabled:  &enabled,
	}

	newUser, err := users.CreateUserWithRole(client, user, "user")
	require.NoError(c.T(), err)

	newUser.Password = user.Password

	standardUserClient, err := client.AsUser(newUser)
	require.NoError(c.T(), err)

	c.standardUserClient = standardUserClient
}

func (c *CustomClusterProvisioningTestSuite) TestProvisioningK3SCustomCluster() {
	nodeRolesAll := []machinepools.NodeRoles{provisioninginput.AllRolesPool}
	nodeRolesShared := []machinepools.NodeRoles{provisioninginput.EtcdControlPlanePool, provisioninginput.WorkerPool}
	nodeRolesDedicated := []machinepools.NodeRoles{provisioninginput.EtcdPool, provisioninginput.ControlPlanePool, provisioninginput.WorkerPool}

	tests := []struct {
		name      string
		client    *rancher.Client
		nodeRoles []machinepools.NodeRoles
	}{
		{"1 Node all roles " + provisioninginput.AdminClientName.String(), c.client, nodeRolesAll},
		{"1 Node all roles " + provisioninginput.StandardClientName.String(), c.standardUserClient, nodeRolesAll},
		{"2 nodes - etcd/cp roles per 1 node " + provisioninginput.AdminClientName.String(), c.client, nodeRolesShared},
		{"2 nodes - etcd/cp roles per 1 node " + provisioninginput.StandardClientName.String(), c.standardUserClient, nodeRolesShared},
		{"3 nodes - 1 role per node " + provisioninginput.AdminClientName.String(), c.client, nodeRolesDedicated},
		{"3 nodes - 1 role per node " + provisioninginput.StandardClientName.String(), c.standardUserClient, nodeRolesDedicated},
	}
	for _, tt := range tests {
		c.clustersConfig.NodesAndRoles = tt.nodeRoles
		permutations.RunTestPermutations(&c.Suite, tt.name, tt.client, c.clustersConfig, permutations.K3SCustomCluster, nil, nil)
	}
}

func (c *CustomClusterProvisioningTestSuite) TestProvisioningK3SCustomClusterDynamicInput() {
	isWindows := false
	for _, pool := range c.clustersConfig.NodesAndRoles {
		if pool.Windows {
			isWindows = true
			break
		}
	}
	require.False(c.T(), isWindows, "Windows nodes are not supported for k3s Clusters")
	if len(c.clustersConfig.NodesAndRoles) == 0 {
		c.T().Skip()
	}

	tests := []struct {
		name   string
		client *rancher.Client
	}{
		{provisioninginput.AdminClientName.String(), c.client},
		{provisioninginput.StandardClientName.String(), c.standardUserClient},
	}
	for _, tt := range tests {
		permutations.RunTestPermutations(&c.Suite, tt.name, tt.client, c.clustersConfig, permutations.K3SCustomCluster, nil, nil)
	}
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestCustomClusterK3SProvisioningTestSuite(t *testing.T) {
	suite.Run(t, new(CustomClusterProvisioningTestSuite))
}
