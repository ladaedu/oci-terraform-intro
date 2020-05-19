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
  "os/exec"
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
	t.Run("checkVCN", checkVCN)
  t.Run("webLoadbalancer", webLoadbalancer)
}

func sshBastion(t *testing.T) {
  hosts := bastionHost(t)
  for i := 0; i < len(hosts); i++ {
    ssh.CheckSshConnection(t, hosts[i])
  }
}

func sshWeb(t *testing.T) {
  bastion_hosts := bastionHost(t)
  web_hosts := webHost(t)
  for bi := 0; bi < len(bastion_hosts); bi++ {
    for wi := 0; wi < len(web_hosts); wi++ {
      jumpSsh(t, "whoami", sshUserName, false, bastion_hosts[bi], web_hosts[wi])
    }
  }
}

func netstatNginx(t *testing.T) {
	netstatService(t, nginxName, nginxPort, 1)
}

func curlWebServer(t *testing.T) {
	curlService(t, "nginx", "", "80", "200")
}

func checkVCN(t *testing.T) {
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
	expected := "Web VCN"
	actual := response.Vcn.DisplayName

	if !strings.HasPrefix(*actual, expected) {
		t.Fatalf("wrong vcn display name: expected %q, got %q", expected, *actual)
	}

	expected = "10.0.0.0/16"
	actual = response.Vcn.CidrBlock

	if expected != *actual {
		t.Fatalf("wrong cidr block: expected %q, got %q", expected, *actual)
	}
}

func sanitizedVcnId(t *testing.T) string {
	raw := terraform.Output(t, options, "VCNID")
	return strings.Split(raw, "\"")[1]
}

func webLoadbalancer(t *testing.T) {
  loadbalancer_ip := terraform.OutputList(t, options, "lb_ip")[0]
  // curl -s -o /dev/null http://132.145.229.197:80/
  command := exec.Command("curl", loadbalancer_ip)
  out, err := command.Output()
  
  if err != nil {
    t.Fatal("Error during curl:", err)
  }
  
  expected := time.Now().Format("02/Jan/2006")
  if !strings.Contains(string(out), expected) {
    t.Fatal("Unexpected answer from load balancer", string(out))
  }
}

// ~~~~~~~~~~~~~~~~ Helper functions ~~~~~~~~~~~~~~~~

func bastionHost(t *testing.T) []ssh.Host {
	bastionIP := terraform.OutputList(t, options, "BastionPublicIP")[0]
	bastionIP_trim := strings.Trim(bastionIP, "[]")
  bastionIP_split := strings.Split(bastionIP_trim, " ")
  sshHosts := make([]ssh.Host, len(bastionIP_split))
  for i := 0; i < len(bastionIP_split); i++ {
    sshHosts[i] = sshHost(t, bastionIP_split[i])
  }
  return sshHosts
}

func webHost(t *testing.T) []ssh.Host {
	webIP := terraform.OutputList(t, options, "WebServerPrivateIPs")[0]
	webIP_trim := strings.Trim(webIP, "[]")
  webIP_split := strings.Split(webIP_trim, " ")
  sshHosts := make([]ssh.Host, len(webIP_split))
  for i := 0; i < len(webIP_split); i++ {
    sshHosts[i] = sshHost(t, webIP_split[i])
  }
  return sshHosts
}

func sshHost(t *testing.T, ip string) ssh.Host {
	return ssh.Host{
		Hostname:    ip,
		SshUserName: sshUserName,
		SshKeyPair:  loadKeyPair(t),
	}
}

func curlService(t *testing.T, serviceName string, path string, port string, returnCode string) {
	bastionHost := bastionHost(t)[0]
	webHosts := webHost(t)

	for _, h := range webHosts {
    cp := h.Hostname
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
	return terraform.OutputList(t, options, "WebServerPrivateIPs")
}

func jumpSsh(t *testing.T, command string, expected string, retryAssert bool, bastionHost ssh.Host, webHost ssh.Host) string {
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
  bastion_hosts := bastionHost(t)
  web_hosts := webHost(t)
	command := fmt.Sprintf("sudo netstat -tnlp | grep '%s' | grep ':%s' | wc -l", service, port)
	jumpSsh(t, command, strconv.Itoa(expectedCount), true, bastion_hosts[0], web_hosts[0])
}
