package terratest

import (
	"context"
	"fmt"
	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/gruntwork-io/terratest/modules/ssh"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/gruntwork-io/terratest/modules/http-helper"
	"github.com/oracle/oci-go-sdk/common"
	"github.com/oracle/oci-go-sdk/core"
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
	maxRetries          = 20
	sleepBetweenRetries = 5 * time.Second
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
	t.Run("checkLoadBalancer", checkLoadBalancer)
	t.Run("checkWebServerShape", checkWebServerShape)
	t.Run("checkBastionSubnet", checkBastionSubnet)
}

func sshBastion(t *testing.T) {
	for i:=0; i<getBastionCount(); i++ {
		fmt.Printf("== Checking Bastion #%d\n", i)
		ssh.CheckSshConnection(t, bastionHost(t, i))
	}
}

func sshWeb(t *testing.T) {
	for i:=0; i<getWebCount(); i++ {
		fmt.Printf("== Checking WebServer #%d\n", i)
		jumpSsh(t, "whoami", sshUserName, false, i)
	}
}

func netstatNginx(t *testing.T) {
	for i:=0; i<getWebCount(); i++ {
		netstatService(t, nginxName, nginxPort, 1, i)
	}
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
	fmt.Printf("VcnID: %s\n", vcnId)
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

func checkLoadBalancer(t *testing.T) {
	lbIP := terraform.OutputList(t, options, "lb_ip")[0]
	webIPs := webServerIPs(t)
	address := fmt.Sprintf("http://%s:80/", lbIP)
	description := fmt.Sprintf("Get LB on %s", address)

	reachedIPs := make([]string, 0)
	for i:=0; i<2*len(webIPs); i++ {
		url := retry.DoWithRetry(t, description, maxRetries, sleepBetweenRetries, func() (string, error) {
			code, body, err := http_helper.HttpGetE(t, address, nil)
			if err != nil {
				return "", err
			}

			if code != 200 {
				return "", fmt.Errorf("Bad http response code: expected %d, got %d", 200, code)
			}

			// first line of output contains IP
			out := strings.Split(body, "\n")[0]
			// line has format "Server address: <ip>:<port>"
			out = strings.Trim(strings.Split(out, ":")[1], " ")
			return out, nil
		})

		if !contains(webIPs, url) {
			t.Fatalf("%q: address %q not contained in %v", description, url, webIPs)
		}

		if !contains(reachedIPs, url) {
			reachedIPs = append(reachedIPs, url)
		}
	}

	for _, ip := range webIPs {
		if !contains(reachedIPs, ip) {
			t.Fatalf("%q: address %q not reached during round robin", description, ip)
		}
	}
}

func checkWebServerShape(t *testing.T) {
	config := common.CustomProfileConfigProvider("", "CzechEdu")
	c, err := core.NewComputeClientWithConfigurationProvider(config)

	if err != nil {
		t.Fatalf("error in creating client: %s", err.Error())
	}

	webIDs := getOutputList(t, "WebServerIDs")
	for _, id := range webIDs {
		request := core.GetInstanceRequest{}
		fmt.Printf("WebID: %s\n", id)
		request.InstanceId = &id

		response, err := c.GetInstance(context.Background(), request)
		if err != nil {
			t.Fatalf("error in calling instance: %s", err.Error())
		}

		expected := "VM.Standard2.1"
		actual := response.Instance.Shape

		if expected != *actual {
			t.Fatalf("wrong VM shape: expected %q, got %q", expected, *actual)
		}

		expected = "eu-frankfurt-1"
		actual = response.Instance.Region

		if (expected != *actual) {
			t.Fatalf("wrong VM region: expected %q, got %q", expected, *actual)
		}
	}
}

func checkBastionSubnet(t *testing.T) {
	config := common.CustomProfileConfigProvider("", "CzechEdu")
	c, err := core.NewVirtualNetworkClientWithConfigurationProvider(config)

	if err != nil {
		t.Fatalf("error in creating client: %s", err.Error())
	}

	bastionSubnetIDs := getOutputList(t, "BastionSubnetIDs")
	for _, id := range bastionSubnetIDs {
		request := core.GetSubnetRequest{}
		fmt.Printf("Subnet ID: %s\n", id)
		request.SubnetId = &id

		response, err := c.GetSubnet(context.Background(), request)
		if err != nil {
			t.Fatalf("error in calling subnet %q: %s", id, err.Error())
		}

		// check CIDR blocks
		expected := []string{"10.0.100.0/28", "10.0.100.16/28", "10.0.100.32/28"}
		actual := response.Subnet.CidrBlock

		if !contains(expected, *actual) {
			t.Fatalf("wrong subnet block: expected one of %v, got %q", expected, *actual)
		}

		// check ingress list
		secListIds := response.Subnet.SecurityListIds
		if len(secListIds) == 0 {
			t.Fatalf("subnet %q has no security lists attached", id)
		}

		for _, sId := range secListIds {
			request2 := core.GetSecurityListRequest{}
			request2.SecurityListId = &sId

			response2, err2 := c.GetSecurityList(context.Background(), request2)
			if err2 != nil {
				t.Fatalf("error in calling security list: %s", err.Error())
			}

			ingressRules := response2.SecurityList.IngressSecurityRules
			if len(ingressRules) == 0 {
				t.Fatalf("security list %q has no ingress rules", sId)
			}

			for _, rule := range ingressRules {
				// only TCP traffic allowed
				if *rule.Protocol != "6" {
					t.Fatalf("wrong protocol version: expected %q, got %q", "6", *rule.Protocol)
				}

				if rule.TcpOptions == nil {
					t.Fatalf("ingress rule has no TCP options")
				}

				tcpOpts := *rule.TcpOptions
				if tcpOpts.DestinationPortRange == nil {
					t.Fatalf("ingress rule has no destination port range")
				}

				portRange := *tcpOpts.DestinationPortRange

				// only SSH allowed
				if portRange.Max == nil || portRange.Min == nil {
					t.Fatalf("a source range bound is unset")
				}
				if *portRange.Max != 22 || *portRange.Min != 22 {
					t.Fatalf("wrong source port range: expected %d-%d, got %d-%d", 22, 22, *portRange.Max, *portRange.Min)
				}
			}
		}
	}
}

func sanitizedVcnId(t *testing.T) string {
	return terraform.OutputList(t, options, "VcnID")[0]
}

// ~~~~~~~~~~~~~~~~ Helper functions ~~~~~~~~~~~~~~~~

func getOutputList(t *testing.T, field string) []string {
	list := terraform.OutputList(t, options, field)[0]
	list = strings.Trim(strings.Trim(list, "["), "]")
	return strings.Split(list, " ")
}

func getBastionCount() int {
	count, err := strconv.Atoi(os.Getenv("TF_VAR_BastionVMCount"))
	if (err != nil) {
		return 1
	}
	return count
}

func getWebCount() int {
	count, err := strconv.Atoi(os.Getenv("TF_VAR_WebVMCount"))
	if (err != nil) {
		return 1
	}
	return count
}

func contains(arr []string, needle string) bool {
	for _, v := range arr {
		if v == needle {
			return true
		}
	}
	return false
}

func bastionHost(t *testing.T, index int) ssh.Host {
	bastionIP := getOutputList(t, "BastionPublicIP")[index]
	return sshHost(t, bastionIP)
}

func webHost(t *testing.T, index int) ssh.Host {
	webIP := getOutputList(t, "WebServerPrivateIPs")[index]
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

	for _, host := range webIPs {
		command := curl(host, port, path)
		description := fmt.Sprintf("curl to %s on %s:%s%s", serviceName, host, port, path)

		out := retry.DoWithRetry(t, description, maxRetries, sleepBetweenRetries, func() (string, error) {
			out, err := ssh.CheckSshCommandE(t, bastionHost, command)
			if err != nil {
				return "", err
			}

			out = strings.TrimSpace(out)
			return out, nil
		})

		if out != returnCode {
			t.Fatalf("%s on %s: expected %q, got %q", serviceName, host, returnCode, out)
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
	bastionHost := bastionHost(t, 0)
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

func netstatService(t *testing.T, service string, port string, expectedCount int, index int) {
	command := fmt.Sprintf("sudo netstat -tnlp | grep '%s' | grep ':%s' | wc -l", service, port)
	jumpSsh(t, command, strconv.Itoa(expectedCount), true, index)
}
