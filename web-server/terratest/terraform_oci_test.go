package terratest

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/gruntwork-io/terratest/modules/ssh"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/oracle/oci-go-sdk/common"
	"github.com/oracle/oci-go-sdk/core"
)

const (
	// OCI params
	sshUserName    = "opc"
	nginxName      = "nginx"
	nginxPath      = "/usr/sbin"
	nginxPort      = "80"
	startupLogFile = "startup-log.txt" //added
	// Terratest retries
	maxRetries          = 20
	sleepBetweenRetries = 5 * time.Second
)

var (
	options        *terraform.Options
	webServerCount int // added
	bastionCount   int // added
)

func terraformEnvOptions(t *testing.T) {
	options = &terraform.Options{
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

	ipMatcher := regexp.MustCompile(`[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+`)

	bastionIPs := terraform.Output(t, options, "BastionPublicIP")
	bastionCount = len(ipMatcher.FindAllStringIndex(bastionIPs, -1))

	webServerIPs := terraform.Output(t, options, "WebServerPrivateIPs")
	webServerCount = len(ipMatcher.FindAllStringIndex(webServerIPs, -1))
}

func TestTerraform(t *testing.T) {
	terraformEnvOptions(t)

	// terraform.WorkspaceSelectOrNew(t, options, "terratest-vita")

	terraform.InitAndApply(t, options)
	runSubtests(t)
	defer terraform.Destroy(t, options)
}

func TestWithoutProvisioning(t *testing.T) {
	terraformEnvOptions(t)

	runSubtests(t)
}

func runSubtests(t *testing.T) {
	t.Run("sshBastion", sshBastion)           // rewritten to check all bastions
	t.Run("sshWeb", sshWeb)                   // rewritten to check all web servers
	t.Run("nginxVersion", nginxVersion)       // rewritten to check all web servers
	t.Run("startupFinished", startupFinished) // rewritten to check all web servers
	t.Run("startupNoErrors", startupNoErrors) // rewritten to check all web servers
	t.Run("netstatNginx", netstatNginx)       // rewritten to send the command to all web servers
	t.Run("curlWebServer", curlWebServer)     // rewritten to check all web servers
	t.Run("loadBalancerCheck", loadBalancer)  // rewritten to check all web servers
	// t.Run("checkVpn", checkVpn)
}

// tests all bastions
func sshBastion(t *testing.T) {
	for i := 0; i < bastionCount; i++ {
		ssh.CheckSshConnection(t, bastionHost(t, i))
	}
}

// tests all web servers through all bastions
func sshWeb(t *testing.T) {
	re := regexp.MustCompile(fmt.Sprintf("^%s$", regexp.QuoteMeta(sshUserName)))

	for indexBastion := 0; indexBastion < bastionCount; indexBastion++ {
		for indexWeb := 0; indexWeb < webServerCount; indexWeb++ {
			jumpSsh(t, indexBastion, indexWeb, "whoami", re, false)
		}
	}
}

// tests all web servers through all bastions
func nginxVersion(t *testing.T) {
	// we do not care about the patch version
	re := regexp.MustCompile(fmt.Sprintf(`nginx version: %s/1\.16\.[0-9]+`, regexp.QuoteMeta(nginxName)))

	for indexBastion := 0; indexBastion < bastionCount; indexBastion++ {
		for indexWeb := 0; indexWeb < webServerCount; indexWeb++ {
			jumpSsh(t, indexBastion, indexWeb, fmt.Sprintf("%s/%s -v", nginxPath, nginxName), re, false)
		}
	}
}

// tests all web servers through all bastions
// checks that the startup script started and finished
func startupFinished(t *testing.T) {
	re := regexp.MustCompile(`.*userdata\.[0-9]+.start.*userdata\.[0-9]+.finish`)

	for indexBastion := 0; indexBastion < bastionCount; indexBastion++ {
		for indexWeb := 0; indexWeb < webServerCount; indexWeb++ {
			jumpSsh(t, indexBastion, indexWeb, `echo ~/userdata.*`, re, false)
		}
	}
}

// tests all web servers through all bastions
// checks that the startup script have not logged any errors
func startupNoErrors(t *testing.T) {
	successString := "No errors"

	re := regexp.MustCompile(fmt.Sprintf(`^%s$`, regexp.QuoteMeta(successString)))

	for indexBastion := 0; indexBastion < bastionCount; indexBastion++ {
		for indexWeb := 0; indexWeb < webServerCount; indexWeb++ {
			jumpSsh(t, indexBastion, indexWeb, fmt.Sprintf(`grep -qe error %s || echo "%s"`, startupLogFile, successString), re, true)
		}
	}
}

func netstatNginx(t *testing.T) {
	netstatService(t, nginxName, nginxPort, 1)
}

// tests connections through each bastion
func curlWebServer(t *testing.T) {
	for indexBastion := 0; indexBastion < bastionCount; indexBastion++ {
		curlService(t, indexBastion, "nginx", "", "80", "200")
	}
}

func checkVpn(t *testing.T) {
	// client
	config := common.CustomProfileConfigProvider("", "DEFAULT")
	c, _ := core.NewVirtualNetworkClientWithConfigurationProvider(config)
	//c, _ := core.NewVirtualNetworkClientWithConfigurationProvider(common.DefaultConfigProvider())
	c.UserAgent = "terratest"

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
	raw := terraform.Output(t, options, "VcnID")

	return strings.Split(raw, "\"")[1]
}

// ~~~~~~~~~~~~~~~~ Helper functions ~~~~~~~~~~~~~~~~

// rewritten to connect to a particular bastion
func bastionHost(t *testing.T, index int) ssh.Host {
	bastionIPs := terraform.OutputList(t, options, "BastionPublicIP")[0]
	bastionIP := strings.Split(bastionIPs[1:len(bastionIPs)-1], " ")[index]

	return sshHost(t, bastionIP)
}

// rewritten to connect to a particular web server
func webHost(t *testing.T, index int) ssh.Host {
	webIPs := terraform.OutputList(t, options, "WebServerPrivateIPs")[0]
	webIP := strings.Split(webIPs[1:len(webIPs)-1], " ")[index]

	return sshHost(t, webIP)
}

func sshHost(t *testing.T, ip string) ssh.Host {
	return ssh.Host{
		Hostname:    ip,
		SshUserName: sshUserName,
		SshKeyPair:  loadKeyPair(t),
	}
}

// checks that the loadbalancer spreads requests evenly
func loadBalancer(t *testing.T) {
	const attempts = 100

	results := make([]int, webServerCount)

	lb_ip := strings.Split(terraform.Output(t, options, "lb_ip"), "\"")[1]

	for i := 0; i < attempts; i++ {
		var out bytes.Buffer

		cmd := exec.Command("curl", "-s", "-w", "%{http_code}", lb_ip)
		cmd.Stdout = &out
		err := cmd.Run()

		if err != nil {
			t.Fatalf("error in calling ssh: %s", err.Error())
		}

		re := regexp.MustCompile("web[0-9]+")
		web_idx, _ := strconv.Atoi(re.FindString(out.String())[3:])

		results[web_idx]++

	}

	var min int = attempts
	var max int = 0

	for i := range results {
		if results[i] < min {
			min = results[i]
		}

		if results[i] > max {
			max = results[i]
		}
	}

	if max-min > 2 {
		t.Fatalf("loadbalancer does not balance enough")
	}
}

// rewritten to connect through a particular bastion
func curlService(t *testing.T, indexBastion int, serviceName string, path string, port string, returnCode string) {
	bastionHost := bastionHost(t, indexBastion)
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

		if !strings.HasPrefix(out, returnCode) {
			t.Fatalf("%s on %s: expected %q, got %q", serviceName, cp, returnCode, out)
		}
	}
}

func curl(host string, port string, path string) string {
	return fmt.Sprintf("curl -s -o /dev/null -w '%%{http_code}' http://%s:%s%s", host, port, path)
}

func webServerIPs(t *testing.T) []string {
	return terraform.OutputList(t, options, "WebServerPrivateIPs")
}

// rewritten to connect through a particular bastion to a particular web server
// expected changed to regexp to add more freedom
func jumpSsh(t *testing.T, indexBastion int, indexWeb int, command string, expected *regexp.Regexp, retryAssert bool) string {
	bastionHost := bastionHost(t, indexBastion)
	webHost := webHost(t, indexWeb)
	description := fmt.Sprintf("ssh jump to %q with command %q", webHost.Hostname, command)

	out := retry.DoWithRetry(t, description, maxRetries, sleepBetweenRetries, func() (string, error) {
		out, err := ssh.CheckPrivateSshConnectionE(t, bastionHost, webHost, command)
		if err != nil {
			return "", err
		}

		out = strings.TrimSpace(out)
		if retryAssert && !expected.MatchString(out) {
			return "", fmt.Errorf("assert with retry: expected %q, got %q", expected, out)
		}
		return out, nil
	})

	if !expected.MatchString(out) {
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

// tests all web servers through each of the bastions
func netstatService(t *testing.T, service string, port string, expectedCount int) {
	command := fmt.Sprintf("sudo netstat -tnlp | grep '%s' | grep ':%s' | wc -l", service, port)
	re := regexp.MustCompile(fmt.Sprintf("^%d$", expectedCount))

	for indexBastion := 0; indexBastion < bastionCount; indexBastion++ {
		for indexWeb := 0; indexWeb < webServerCount; indexWeb++ {
			jumpSsh(t, indexBastion, indexWeb, command, re, true)
		}
	}
}
