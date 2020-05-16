package terratest

import (
	"context"
	"fmt"
	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/gruntwork-io/terratest/modules/ssh"
	"github.com/gruntwork-io/terratest/modules/terraform"
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
	t.Run("checkVcn", checkVcn)
	t.Run("webServerOut", webServerOut)
	t.Run("loadBalancerRuns", loadBalancerRuns)
    t.Run("loadBalancerRoundRobin", loadBalancerRoundRobin)
    t.Run("noDirectSshToWebServer", noDirectSshToWebServer)
}

// Test SSH to Bastion
func sshBastion(t *testing.T) {
	ssh.CheckSshConnection(t, bastionHost(t))
}

// Test that we cannot use SSH directly to web servers
func noDirectSshToWebServer(t *testing.T) {
    for i := 0; i < numberOfWebServers(t); i++ {
        _, err := ssh.CheckSshCommandE(t, webHost(t, i), "echo foo")

        if err == nil {
            fmt.Errorf("direct SSH to web server worked: %q", err)
        }
    }
}

// Test jump SSH to every web server
func sshWeb(t *testing.T) {
    for i := 0; i < numberOfWebServers(t); i++ {
        jumpSsh(t, i, "whoami", sshUserName, false)
    }
}

// Test web server running for every web server
func netstatNginx(t *testing.T) {
    for i := 0; i < numberOfWebServers(t); i++ {
        netstatService(t, i, nginxName, nginxPort, 1)
    }
}

// Test web server responding correctly for every web server
func curlWebServer(t *testing.T) {
    curlService(t, "nginx", "", "80", "200")
}

// Test VCN
func checkVcn(t *testing.T) {
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

// Test that web server has access outside of our system
func webServerOut(t *testing.T) {
    for i := 0; i < numberOfWebServers(t); i++ {
        jumpSsh(t, i, "if wget -q --spider 'http://google.com'; then echo true; else echo false; fi", "true", true)
    }
}

// Test that load balancer is available
func loadBalancerRuns(t *testing.T) {
    bastionHost := bastionHost(t)
    loadBalancerIP := terraform.OutputList(t, options, "lb_ip")[0]

    command := curl_http_code(loadBalancerIP, "80", "")
    description := fmt.Sprintf("curl to load balancer")

    out := retry.DoWithRetry(t, description, maxRetries, sleepBetweenRetries, func() (string, error) {
        out, err := ssh.CheckSshCommandE(t, bastionHost, command)
        if err != nil {
            return "", err
        }

        out = strings.TrimSpace(out)
        return out, nil
    })

    if !strings.Contains(out, "200") {
        t.Fatalf("curl on load balancer: expected 200, got %q", out)
    }
}

// Test that load balancer really balances load
func loadBalancerRoundRobin(t *testing.T) {
    previousServers := ""
    for i := 0; i < numberOfWebServers(t); i++ {
        currentServer := loadBalancerRoundRobinIteration(t)
        if strings.Contains(previousServers, currentServer) {
            t.Fatalf("round robin repeated server: %s", currentServer)
        }
        previousServers = previousServers + " " + currentServer
    }
}

func sanitizedVcnId(t *testing.T) string {
	raw := terraform.Output(t, options, "VcnID")
	return strings.Split(raw, "\"")[1]
}

// ~~~~~~~~~~~~~~~~ Helper functions ~~~~~~~~~~~~~~~~

func numberOfWebServers(t *testing.T) int {
	webIPs := terraform.OutputList(t, options, "WebServerPrivateIPs")[0]
	return len(strings.Split(webIPs, " "))
}

func loadBalancerRoundRobinIteration(t *testing.T) string {
    bastionHost := bastionHost(t)
    loadBalancerIP := terraform.OutputList(t, options, "lb_ip")[0]

    command := curl_server_name(loadBalancerIP)
    description := fmt.Sprintf("curl to load balancer")

    out := retry.DoWithRetry(t, description, maxRetries, sleepBetweenRetries, func() (string, error) {
        out, err := ssh.CheckSshCommandE(t, bastionHost, command)
        if err != nil {
            return "", err
        }

        out = strings.TrimSpace(out)
        return out, nil
    })

    return out
}

func bastionHost(t *testing.T) ssh.Host {
	bastionIP := terraform.OutputList(t, options, "BastionPublicIP")[0]
	return sshHost(t, bastionIP)
}

func webHost(t *testing.T, webServerID int) ssh.Host {
	webIPs := terraform.OutputList(t, options, "WebServerPrivateIPs")[0]
	webIP := strings.Split(strings.Trim(strings.Trim(webIPs, "["), "]"), " ")[webServerID]
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
	bastionHost := bastionHost(t)
	webIPs := webServerIPs(t)

	for _, cp := range webIPs {
		re := strings.NewReplacer("[", "", "]", "")
		host := re.Replace(cp)
		command := curl_http_code(host, port, path)
		description := fmt.Sprintf("curl to %s on %s:%s%s", serviceName, cp, port, path)

		out := retry.DoWithRetry(t, description, maxRetries, sleepBetweenRetries, func() (string, error) {
			out, err := ssh.CheckSshCommandE(t, bastionHost, command)
			if err != nil {
				return "", err
			}

			out = strings.TrimSpace(out)
			return out, nil
		})

		if !strings.Contains(out, returnCode) {
			t.Fatalf("%s on %s: expected %q, got %q", serviceName, cp, returnCode, out)
		}
	}
}

func curl_http_code(host string, port string, path string) string {
	return fmt.Sprintf("curl -s -o /dev/null -w '%%{http_code}' http://%s:%s%s", host, port, path)
}

func curl_server_name(host string) string {
	return fmt.Sprintf("curl -s http://%s |grep 'Server name' |sed 's/Server name: //'", host)
}

func webServerIPs(t *testing.T) []string {
	return terraform.OutputList(t, options, "WebServerPrivateIPs")
}

func jumpSsh(t *testing.T, webServerID int, command string, expected string, retryAssert bool) string {
	bastionHost := bastionHost(t)
	webHost := webHost(t, webServerID)
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

func netstatService(t *testing.T, webServerID int, service string, port string, expectedCount int) {
	command := fmt.Sprintf("sudo netstat -tnlp | grep '%s' | grep ':%s' | wc -l", service, port)
	jumpSsh(t, webServerID, command, strconv.Itoa(expectedCount), true)
}
