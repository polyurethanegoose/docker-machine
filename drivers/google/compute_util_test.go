package google

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	raw "google.golang.org/api/compute/v1"
	"regexp"
)

func TestDefaultTag(t *testing.T) {
	tags := parseTags(&Driver{Tags: ""})

	assert.Equal(t, []string{"docker-machine"}, tags)
}

func TestAdditionalTag(t *testing.T) {
	tags := parseTags(&Driver{Tags: "tag1"})

	assert.Equal(t, []string{"docker-machine", "tag1"}, tags)
}

func TestAdditionalTags(t *testing.T) {
	tags := parseTags(&Driver{Tags: "tag1,tag2"})

	assert.Equal(t, []string{"docker-machine", "tag1", "tag2"}, tags)
}

func TestPortsUsed(t *testing.T) {
	var tests = []struct {
		description   string
		computeUtil   *ComputeUtil
		expectedPorts []string
		expectedError error
	}{
		{"use docker port", &ComputeUtil{}, []string{"2376/tcp"}, nil},
		{"use docker and swarm port", &ComputeUtil{SwarmMaster: true, SwarmHost: "tcp://host:3376"}, []string{"2376/tcp", "3376/tcp"}, nil},
		{"use docker and non default swarm port", &ComputeUtil{SwarmMaster: true, SwarmHost: "tcp://host:4242"}, []string{"2376/tcp", "4242/tcp"}, nil},
		{"include additional ports", &ComputeUtil{openPorts: []string{"80", "2377/udp"}}, []string{"2376/tcp", "80/tcp", "2377/udp"}, nil},
	}

	for _, test := range tests {
		ports, err := test.computeUtil.portsUsed()

		assert.Equal(t, test.expectedPorts, ports)
		assert.Equal(t, test.expectedError, err)
	}
}

func TestMissingOpenedPorts(t *testing.T) {
	var tests = []struct {
		description     string
		allowed         []*raw.FirewallAllowed
		ports           []string
		expectedMissing map[string][]string
	}{
		{"no port opened", []*raw.FirewallAllowed{}, []string{"2376"}, map[string][]string{"tcp": {"2376"}}},
		{"docker port opened", []*raw.FirewallAllowed{{IPProtocol: "tcp", Ports: []string{"2376"}}}, []string{"2376"}, map[string][]string{}},
		{"missing swarm port", []*raw.FirewallAllowed{{IPProtocol: "tcp", Ports: []string{"2376"}}}, []string{"2376", "3376"}, map[string][]string{"tcp": {"3376"}}},
		{"missing docker port", []*raw.FirewallAllowed{{IPProtocol: "tcp", Ports: []string{"3376"}}}, []string{"2376", "3376"}, map[string][]string{"tcp": {"2376"}}},
		{"both ports opened", []*raw.FirewallAllowed{{IPProtocol: "tcp", Ports: []string{"2376", "3376"}}}, []string{"2376", "3376"}, map[string][]string{}},
		{"more ports opened", []*raw.FirewallAllowed{{IPProtocol: "tcp", Ports: []string{"2376", "3376", "22", "1024-2048"}}}, []string{"2376", "3376"}, map[string][]string{}},
		{"additional missing", []*raw.FirewallAllowed{{IPProtocol: "tcp", Ports: []string{"2376", "2377/tcp"}}}, []string{"2377/udp", "80/tcp", "2376"}, map[string][]string{"tcp": {"80"}, "udp": {"2377"}}},
	}

	for _, test := range tests {
		firewall := &raw.Firewall{Allowed: test.allowed}

		missingPorts := missingOpenedPorts(firewall, test.ports)

		assert.Equal(t, test.expectedMissing, missingPorts, test.description)
	}
}

func TestNetworkConfiguration(t *testing.T) {
	driver := &Driver{
		Project:           "project",
		Zone:              "zone-a",
		Network:           "network",
		Subnetwork:        "subnetwork",
		UseInternalIPOnly: false,
	}

	cu := &ComputeUtil{
		project:           driver.Project,
		zone:              driver.Zone,
		network:           driver.Network,
		subnetwork:        driver.Subnetwork,
		useInternalIPOnly: driver.UseInternalIPOnly,
		globalURL:         "https://global",
	}

	instance := &raw.Instance{}
	cu.prepareNetworkingInterfaces(driver, instance)

	require.Len(t, instance.NetworkInterfaces, 1)

	assert.Equal(t, "https://global/networks/network", instance.NetworkInterfaces[0].Network)
	assert.Equal(t, "projects/project/regions/zone/subnetworks/subnetwork", instance.NetworkInterfaces[0].Subnetwork)

	assert.Len(t, instance.NetworkInterfaces[0].AccessConfigs, 1)
	assert.Equal(t, instance.NetworkInterfaces[0].AccessConfigs[0], &raw.AccessConfig{Type: "ONE_TO_ONE_NAT"})
}

func TestNetworkConfigurationWithAdditionalNetworks(t *testing.T) {
	driver := &Driver{
		Project:            "project",
		Zone:               "zone-a",
		Network:            "network",
		Subnetwork:         "subnetwork",
		UseInternalIPOnly:  false,
		AdditionalNetworks: "network-1:subnetwork-1,network-2:,network-3:subnetwork-3",
	}

	cu := &ComputeUtil{
		project:            driver.Project,
		zone:               driver.Zone,
		network:            driver.Network,
		subnetwork:         driver.Subnetwork,
		additionalNetworks: driver.AdditionalNetworks,
		useInternalIPOnly:  driver.UseInternalIPOnly,
		globalURL:          "https://global",
	}

	instance := &raw.Instance{}
	err := cu.prepareNetworkingInterfaces(driver, instance)

	require.NoError(t, err)
	require.Len(t, instance.NetworkInterfaces, 4)

	assert.Equal(t, "https://global/networks/network", instance.NetworkInterfaces[0].Network)
	assert.Equal(t, "projects/project/regions/zone/subnetworks/subnetwork", instance.NetworkInterfaces[0].Subnetwork)
	assert.Len(t, instance.NetworkInterfaces[0].AccessConfigs, 1)
	assert.Equal(t, instance.NetworkInterfaces[0].AccessConfigs[0], &raw.AccessConfig{Type: "ONE_TO_ONE_NAT"})

	assert.Equal(t, "https://global/networks/network-1", instance.NetworkInterfaces[1].Network)
	assert.Equal(t, "projects/project/regions/zone/subnetworks/subnetwork-1", instance.NetworkInterfaces[1].Subnetwork)
	assert.Empty(t, instance.NetworkInterfaces[1].AccessConfigs)

	assert.Equal(t, "https://global/networks/network-2", instance.NetworkInterfaces[2].Network)
	assert.Empty(t, instance.NetworkInterfaces[2].Subnetwork)
	assert.Empty(t, instance.NetworkInterfaces[2].AccessConfigs)

	assert.Equal(t, "https://global/networks/network-3", instance.NetworkInterfaces[3].Network)
	assert.Equal(t, "projects/project/regions/zone/subnetworks/subnetwork-3", instance.NetworkInterfaces[3].Subnetwork)
	assert.Empty(t, instance.NetworkInterfaces[3].AccessConfigs)
}

func TestNetworkConfigurationWithInvalidAdditionalNetwork(t *testing.T) {
	additionalNetworks := []string{
		":",
		"1-network:",
		"network:1-subnetwork",
		"network-",
		":subnetwork-",
	}

	for _, additionalNetwork := range additionalNetworks {
		t.Run(additionalNetwork, func(t *testing.T) {
			driver := &Driver{
				Project:            "project",
				Zone:               "zone-a",
				Network:            "network",
				Subnetwork:         "subnetwork",
				UseInternalIPOnly:  false,
				AdditionalNetworks: additionalNetwork,
			}

			cu := &ComputeUtil{
				project:            driver.Project,
				zone:               driver.Zone,
				network:            driver.Network,
				subnetwork:         driver.Subnetwork,
				additionalNetworks: driver.AdditionalNetworks,
				useInternalIPOnly:  driver.UseInternalIPOnly,
				globalURL:          "https://global",
			}

			instance := &raw.Instance{}
			err := cu.prepareNetworkingInterfaces(driver, instance)

			require.Error(t, err, "Should got an error, that additional network definition is invalid")
		})
	}
}

func TestNetworkAndSubnetworkRegexp(t *testing.T) {
	networkDefinitionRegexp, err := regexp.Compile(networkAndSubnetworkRegexp)
	require.NoError(t, err)

	examples := map[string]bool{
		// only network (valid or invalid)
		"a":   true,
		"a9":  true,
		"aa":  true,
		"a-9": true,
		"9":   false,
		"9a":  false,
		"9-a": false,
		"a-":  false,
		"-":   false,
		"abcdedhfks-asifje843sxijosijdfo9abcdedhfks-asifje843sxijosijdfo9":  true,
		"abcdedhfks-asifje843sxijosijdfo9abcdedhfks-asifje843sxijosijdfo9a": false,
		"abcdedhfks-asifje843s;josijdfo9abcdedhfks-asifje843sxijosijdfo9a":  false,
		// invalid network and subnetwork
		"9a:":  false,
		"9a:a": false,
		// valid network and subnetwork (valid or invalid)
		"a:":    true,
		"a:a":   true,
		"a:a9":  true,
		"a:aa":  true,
		"a:a-9": true,
		"a:9":   false,
		"a:9a":  false,
		"a:9-a": false,
		"a:a-":  false,
		"a:-":   false,
		"a:abcdedhfks-asifje843sxijosijdfo9abcdedhfks-asifje843sxijosijdfo9":  true,
		"a:abcdedhfks-asifje843sxijosijdfo9abcdedhfks-asifje843sxijosijdfo9a": false,
		"a:abcdedhfks-asifje843s;josijdfo9abcdedhfks-asifje843sxijosijdfo9a":  false,
	}

	for network, shouldMatch := range examples {
		if shouldMatch {
			assert.Regexp(t, networkDefinitionRegexp, network)
		} else {
			assert.NotRegexp(t, networkDefinitionRegexp, network)
		}
	}
}
