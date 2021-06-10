package terratest

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/gruntwork-io/terratest/modules/ssh"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/oracle/oci-go-sdk/v41/common"
	"github.com/oracle/oci-go-sdk/v41/core"
	"github.com/oracle/oci-go-sdk/v41/loadbalancer"
)

const (
	// OCI params
	sshUserName = "opc"
	nginxName   = "nginx"
	nginxPort   = "80"
	// Terratest retries
	maxRetries          = 20
	sleepBetweenRetries = 5 * time.Second
)

var (
	options         *terraform.Options
	bastionsCount   int
	webServersCount int
)

func terraformEnvOptions() *terraform.Options {
	return &terraform.Options{
		TerraformDir: "..",
		Vars: map[string]interface{}{
			"region":           os.Getenv("TF_VAR_region"),
			"tenancy_ocid":     os.Getenv("TF_VAR_tenancy_ocid"),
			"user_ocid":        os.Getenv("TF_VAR_user_ocid"),
			"CompartmentOCID":  os.Getenv("TF_VAR_CompartmentOCID"),
			"fingerprint":      os.Getenv("TF_VAR_fingerprint"),
			"private_key_path": os.Getenv("TF_VAR_private_key_path"),
			// "pass_phrase":      oci.GetPassPhraseFromEnvVar(),
			"ssh_public_key":  os.Getenv("TF_VAR_ssh_public_key"),
			"ssh_private_key": os.Getenv("TF_VAR_ssh_private_key"),
		},
	}
}

func initializeVariables(t *testing.T) {
	bastionsCount = len(getOutputList(t, "BastionPublicIP"))
	webServersCount = len(getOutputList(t, "WebServerPrivateIPs"))
}

func TestTerraform(t *testing.T) {
	options = terraformEnvOptions()

	defer terraform.Destroy(t, options)
	// terraform.WorkspaceSelectOrNew(t, options, "terratest-vita")
	terraform.InitAndApply(t, options)

	runSubtests(t)
}

func TestWithoutProvisioning(t *testing.T) {
	options = terraformEnvOptions()

	runSubtests(t)
}

func runSubtests(t *testing.T) {
	initializeVariables(t)
	t.Run("testInstancesCounts", testInstancesCounts)
	t.Run("sshBastion", sshBastion)
	t.Run("sshWeb", sshWeb)
	t.Run("netstatNginx", netstatNginx)
	t.Run("curlWebServer", curlWebServer)
	t.Run("testLoadBalancerProperties", testLoadBalancerProperties)
	t.Run("testLoadBalancing", testLoadBalancing)
	t.Run("testWebServerProperites", testWebServerProperties)
	t.Run("checkVpn", checkVpn)
}

func exitIfNotNill(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func testWebServerProperties(t *testing.T) {
	webServersIds := getOutputList(t, "WebServerIds")
	availDomains := map[string]bool{}

	for i := 0; i < len(webServersIds); i++ {
		c, err := core.NewComputeClientWithConfigurationProvider(common.DefaultConfigProvider())

		exitIfNotNill(err)

		request := core.GetInstanceRequest{
			InstanceId: &webServersIds[i],
		}

		res, err := c.GetInstance(context.Background(), request)

		exitIfNotNill(err)

		expectedShape := "VM.Standard2.1"
		actualShape := res.Shape

		if expectedShape != *actualShape {
			t.Fatalf("Incorrect compute instance shape. Expected: %s , Actual: %s", expectedShape, *actualShape)
		}

		if _, ok := availDomains[*res.AvailabilityDomain]; ok {

		} else {
			availDomains[*res.AvailabilityDomain] = true
		}
	}

	expectedDomainsUsed := 2
	actualceDomainsUsed := len(availDomains)

	if expectedDomainsUsed != actualceDomainsUsed {
		t.Fatalf("Incorrect number of availability domains used by web servers. Expected: %d , Actual: %d", expectedDomainsUsed, actualceDomainsUsed)
	}
}

func testInstancesCounts(t *testing.T) {
	expectedBastions := 2
	expectedWebServers := 2

	if expectedBastions != bastionsCount {
		t.Fatalf("Incorrect bastions count. Expected: %d , Actual: %d", expectedBastions, bastionsCount)
	}

	if expectedWebServers != webServersCount {
		t.Fatalf("Incorrect bastions count. Expected: %d , Actual: %d", expectedWebServers, webServersCount)
	}
}

func testLoadBalancing(t *testing.T) {

	lbIp := getFromOutputList(t, "lb_ip", 0)
	lbAddress := fmt.Sprintf("http://%s", lbIp)

	servers := []string{"web0", "web1"}
	lastFound := -1

	for i := 0; i < 20; i++ {
		res, err := http.Get(lbAddress)
		exitIfNotNill(err)

		defer res.Body.Close()

		body, err := io.ReadAll(res.Body)

		exitIfNotNill(err)

		strBody := string(body)

		for i := 0; i < len(servers); i++ {
			found := strings.Index(strBody, servers[i])

			if found != -1 {
				fmt.Printf("Received response from %s\n", servers[i])
				if lastFound == i {
					t.Fatalf("Load balancer is not load balancing wit round robin policy.")
				}

				lastFound = found
			}
		}
	}
}

func testLoadBalancerProperties(t *testing.T) {
	client, err := loadbalancer.NewLoadBalancerClientWithConfigurationProvider(common.DefaultConfigProvider())
	ctx := context.Background()

	exitIfNotNill(err)

	lbId := getFromOutputList(t, "LoadBalancerId", 0)

	getLBRequest := loadbalancer.GetLoadBalancerRequest{
		LoadBalancerId: &lbId,
	}

	res, err := client.GetLoadBalancer(ctx, getLBRequest)

	exitIfNotNill(err)

	expectedShape := "100Mbps"
	actualShape := res.ShapeName

	if expectedShape != *actualShape {
		t.Fatalf("Load balancer has unexpected shape. Expected: %s Actual: %s", expectedShape, *actualShape)
	}

	bsName := getFromOutputList(t, "BackendSetName", 0)

	bs := res.BackendSets[bsName]

	expectedPolicy := "ROUND_ROBIN"
	actualPolicy := bs.Policy

	if expectedPolicy != *actualPolicy {
		t.Fatalf("Load balancer backend set has unexpected policy. Expected: %s Actual: %s", expectedPolicy, *actualPolicy)
	}

	// Check that the set contains all the web servers
	expectedBackendsCount := webServersCount
	actualBackendsCount := len(bs.Backends)

	if expectedBackendsCount != actualBackendsCount {
		t.Fatalf("Load balancer backend set size. Expected: %d Actual: %d", expectedBackendsCount, actualBackendsCount)
	}
}

func sshBastion(t *testing.T) {
	for i := 0; i < bastionsCount; i++ {
		fmt.Printf("Checking SSH connection for bastion num %d \n", i)
		ssh.CheckSshConnection(t, bastionHost(t, i))
	}
}

func sshWeb(t *testing.T) {
	for bastionNum := 0; bastionNum < bastionsCount; bastionNum++ {
		for webServerNum := 0; webServerNum < webServersCount; webServerNum++ {
			fmt.Printf("Jump SSH from bastion %d to web server %d", bastionNum, webServerNum)
			jumpSsh(t, "whoami", sshUserName, false, bastionNum, webServerNum)
		}
	}
}

func netstatNginx(t *testing.T) {
	for bastionNum := 0; bastionNum < bastionsCount; bastionNum++ {
		for webServerNum := 0; webServerNum < webServersCount; webServerNum++ {
			fmt.Printf("Checking ngnix on webserver num %d via bastion num %d", webServerNum, bastionNum)
			netstatService(t, nginxName, nginxPort, 1, bastionNum, webServerNum)
		}
	}
}

func curlWebServer(t *testing.T) {
	curlService(t, "nginx", "", "80", "200")
}

func checkVpn(t *testing.T) {
	// client
	// config := common.CustomProfileConfigProvider("", "CzechEdu")
	// c, _ := core.NewVirtualNetworkClientWithConfigurationProvider(config)
	c, _ := core.NewVirtualNetworkClientWithConfigurationProvider(common.DefaultConfigProvider())

	// request
	request := core.GetVcnRequest{}
	vcnId := sanitizedVcnId(t)
	request.VcnId = &vcnId

	// response
	response, err := c.GetVcn(context.Background(), request)

	if err != nil {
		t.Fatalf("error in calling vcn: %s", err.Error())
	}

	// assertions
	expected := "Web VCN-default"
	actual := response.Vcn.DisplayName

	if expected != *actual {
		t.Fatalf("wrong vcn display name: expected %q, got %q", expected, *actual)
	}

	expected = "10.0.0.0/16"
	actual = response.Vcn.CidrBlock

	if expected != *actual {
		t.Fatalf("wrong cidr block: expected %q, got %q", expected, *actual)
	}
}

func sanitizedVcnId(t *testing.T) string {
	raw := terraform.OutputList(t, options, "VcnID")[0]
	return raw
}

// ~~~~~~~~~~~~~~~~ Helper functions ~~~~~~~~~~~~~~~~

func bastionHost(t *testing.T, index int) ssh.Host {
	bastionIP := getFromOutputList(t, "BastionPublicIP", index)

	return sshHost(t, bastionIP)
}

func getOutputList(t *testing.T, outputName string) []string {
	output := terraform.OutputList(t, options, outputName)[0]
	output = strings.TrimLeft(output, "[")
	output = strings.TrimRight(output, "]")
	list := strings.Split(output, " ")
	return list
}

func getFromOutputList(t *testing.T, outputName string, index int) string {
	return getOutputList(t, outputName)[index]
}

func webHost(t *testing.T, index int) ssh.Host {
	webIP := getFromOutputList(t, "WebServerPrivateIPs", index)
	return sshHost(t, webIP)
}

func sshHost(t *testing.T, ip string) ssh.Host {
	return ssh.Host{
		Hostname:    ip,
		SshUserName: sshUserName,
		SshKeyPair:  loadKeyPair(t),
	}
}

func curlService(t *testing.T, serviceName string, path string, port string, returnCode string) {
	bastionHost := bastionHost(t, 0)
	webIPs := webServerIPs(t)

	for _, cp := range webIPs {
		re := strings.NewReplacer("[", "", "]", "")
		host := re.Replace(cp)
		command := curl(host, port, path)
		description := fmt.Sprintf("curl to %s on %s:%s%s", serviceName, cp, port, path)

		out := retry.DoWithRetry(t, description, maxRetries, sleepBetweenRetries, func() (string, error) {
			out, err := ssh.CheckSshCommandE(t, bastionHost, command)
			if err != nil {
				return "", err
			}

			out = strings.TrimSpace(out)
			return out, nil
		})

		if out != returnCode {
			t.Fatalf("%s on %s: expected %q, got %q", serviceName, cp, returnCode, out)
		}
	}
}

func curl(host string, port string, path string) string {
	return fmt.Sprintf("curl -s -o /dev/null -w '%%{http_code}' http://%s:%s%s", host, port, path)
}

func webServerIPs(t *testing.T) []string {
	return getOutputList(t, "WebServerPrivateIPs")
}

func jumpSsh(t *testing.T, command string, expected string, retryAssert bool, fromBastionNum int, toWebServerNum int) string {
	bastionHost := bastionHost(t, fromBastionNum)

	webHost := webHost(t, toWebServerNum)
	description := fmt.Sprintf("ssh jump to %q with command %q", webHost.Hostname, command)

	out := retry.DoWithRetry(t, description, maxRetries, sleepBetweenRetries, func() (string, error) {
		out, err := ssh.CheckPrivateSshConnectionE(t, bastionHost, webHost, command)
		if err != nil {
			return "", err
		}

		out = strings.TrimSpace(out)
		if retryAssert && out != expected {
			return "", fmt.Errorf("assert with retry: expected %q, got %q", expected, out)
		}
		return out, nil
	})

	if out != expected {
		t.Fatalf("command %q on %s: expected %q, got %q", command, webHost.Hostname, expected, out)
	}

	return out
}

func loadKeyPair(t *testing.T) *ssh.KeyPair {
	publicKeyPath := options.Vars["ssh_public_key"].(string)
	publicKey, err := ioutil.ReadFile(publicKeyPath)
	if err != nil {
		t.Fatal(err)
	}

	privateKeyPath := options.Vars["ssh_private_key"].(string)
	privateKey, err := ioutil.ReadFile(privateKeyPath)
	if err != nil {
		t.Fatal(err)
	}

	return &ssh.KeyPair{
		PublicKey:  string(publicKey),
		PrivateKey: string(privateKey),
	}
}

func netstatService(t *testing.T, service string, port string, expectedCount int, fromBastionNum int, toWebServerNum int) {
	command := fmt.Sprintf("sudo netstat -tnlp | grep '%s' | grep ':%s' | wc -l", service, port)
	jumpSsh(t, command, strconv.Itoa(expectedCount), true, fromBastionNum, toWebServerNum)
}
