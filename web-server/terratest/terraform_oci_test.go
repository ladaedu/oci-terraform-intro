package terratest

import (
	"context"
	"fmt"
	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/gruntwork-io/terratest/modules/ssh"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/oracle/oci-go-sdk/common"
	"github.com/oracle/oci-go-sdk/core"
	"github.com/oracle/oci-go-sdk/identity"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

const (
	// OCI params
	sshUserName = "opc"
	nginxName   = "nginx"
	nginxPort   = "80"
	// Terratest retries
	maxRetries          = 3
	sleepBetweenRetries = 3 * time.Second
)

var (
	options *terraform.Options
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
	t.Run("sshBastion", sshBastion)
	t.Run("sshWeb", sshWeb)
	t.Run("netstatNginx", netstatNginx)
	t.Run("curlWebServer", curlWebServer)
	t.Run("checkVpn", checkVpn)
	t.Run("checkAvailabilityDomain", checkAvailabilityDomain)
	t.Run("checkLoadBalancer", checkLoadBalancer)
	t.Run("checkIsWebServerPrivate", checkIsWebServerPrivate)
	t.Run("checkLoadBalancerIsPublic", checkLoadBalancerIsPublic)
	t.Run("checkHostnames", checkHostnames)
}

func sshBastion(t *testing.T) {
	ssh.CheckSshConnection(t, bastionHost(t))
}

func sshWeb(t *testing.T) {
	jumpSsh(t, "whoami", sshUserName, false, 0)
}

func netstatNginx(t *testing.T) {
	netstatService(t, nginxName, nginxPort, 1)
}

func curlWebServer(t *testing.T) {
	curlService(t, "nginx", "", "80", "200")
}

func checkVpn(t *testing.T) {
	// client
	config := common.CustomProfileConfigProvider("", "CzechEdu")
	c, _ := core.NewVirtualNetworkClientWithConfigurationProvider(config)
	// c, _ := core.NewVirtualNetworkClientWithConfigurationProvider(common.DefaultConfigProvider())

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

func checkAvailabilityDomain(t *testing.T) {
	config := common.CustomProfileConfigProvider("", "CzechEdu")
	client, err := identity.NewIdentityClientWithConfigurationProvider(config)

	compartmentID := options.Vars["CompartmentOCID"].(string)
	request := identity.ListAvailabilityDomainsRequest{CompartmentId: &compartmentID}
	response, err := client.ListAvailabilityDomains(context.Background(), request)
	if err != nil {
		t.Fatalf("error in: %s", err.Error())
	}

	expected := "NoND:EU-FRANKFURT-1-AD-1"
	actual := response.Items[0].Name

	if expected != *actual {
		t.Fatalf("wrong availability domain: expected %q, got %q", expected, *actual)
	}
}

func checkLoadBalancer(t *testing.T) {
	lb_address := terraform.OutputList(t, options, "lb_ip")[0]
	command := curl(lb_address, "80", "")
	description := fmt.Sprintf("curl to load balancer on %s:80", lb_address)	

	out := retry.DoWithRetry(t, description, maxRetries, sleepBetweenRetries, func() (string, error) {
		out, err := ssh.CheckSshCommandE(t, bastionHost(t), command)
		if err != nil {
			return "", err
		}

		out = strings.TrimSpace(out)
		return out, nil
	})

	if out != "200" {
		t.Fatalf("can't connect to load balaner on address %s", lb_address)
	}
}

func checkLoadBalancerIsPublic(t *testing.T) {
	lb_is_public := terraform.OutputList(t, options, "lb_is_public")[0]

	// assertions
	expected := "true"
	actual := lb_is_public

	if expected != actual {
		t.Fatalf("Load balaner must be public: expected %q, got %q", expected, actual)
	}
}

func checkIsWebServerPrivate(t *testing.T) {
	webServerDomain := getOutputList(t, "WebServerDomain")[0]

	// assertions
	expected := "private.demo.oraclevcn.com"
	actual := webServerDomain

	if expected != actual {
		t.Fatalf("Web server must be private: expected domain %q, got %q", expected, actual)
	}
}

func checkHostnames(t *testing.T) {
	hostnames := getOutputList(t, "WebServerHostNames")
	
	if len(hostnames) != 3 {
		t.Fatalf("3 hostnames should have been present")
	}

	if hostnames[0] != "web0" {
		t.Fatalf("First hostname should be web0")
	}

	if hostnames[1] != "web1" {
		t.Fatalf("First hostname should be web0")
	}

	if hostnames[2] != "web2" {
		t.Fatalf("First hostname should be web0")
	}
}

func sanitizedVcnId(t *testing.T) string {
	raw := terraform.Output(t, options, "VcnID")
	return strings.Split(raw, "\"")[1]
}

// ~~~~~~~~~~~~~~~~ Helper functions ~~~~~~~~~~~~~~~~

func bastionHost(t *testing.T) ssh.Host {
	bastionIP := terraform.OutputList(t, options, "BastionPublicIP")[0]
	return sshHost(t, bastionIP)
}

func webHost(t *testing.T, index int) ssh.Host {
	webIP := getOutputList(t, "WebServerPrivateIPs")[index]
	return sshHost(t, webIP)
}

func getOutputList(t *testing.T, listName string) []string {
	listStr := terraform.OutputList(t, options, listName)[0]
	listStr = strings.Trim(strings.Trim(listStr, "["), "]")
	list := strings.Split(listStr, " ")
	return list
}

func sshHost(t *testing.T, ip string) ssh.Host {
	return ssh.Host{
		Hostname:    ip,
		SshUserName: sshUserName,
		SshKeyPair:  loadKeyPair(t),
	}
}

func curlService(t *testing.T, serviceName string, path string, port string, returnCode string) {
	bastionHost := bastionHost(t)
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

func jumpSsh(t *testing.T, command string, expected string, retryAssert bool, index int) string {
	bastionHost := bastionHost(t)
	webHost := webHost(t, index)
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

func netstatService(t *testing.T, service string, port string, expectedCount int) {
	command := fmt.Sprintf("sudo netstat -tnlp | grep '%s' | grep ':%s' | wc -l", service, port)
	jumpSsh(t, command, strconv.Itoa(expectedCount), true, 0)
}
