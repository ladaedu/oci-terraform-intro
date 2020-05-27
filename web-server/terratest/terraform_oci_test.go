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
	"os/exec"
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
	t.Run("checkCompute", checkCompute)
	t.Run("checkLoadBalancer", checkLoadBalancer)
	t.Run("sshLoadBalancer", sshLoadBalancer)
	t.Run("curlLoadBalancer", curlLoadBalancer)
}

func sshBastion(t *testing.T) {
	ssh.CheckSshConnection(t, bastionHost(t))
}

func sshWeb(t *testing.T) {
	jumpSsh(t, "whoami", sshUserName, false)
}

func netstatNginx(t *testing.T) {
	netstatService(t, nginxName, nginxPort, 1)
}

func curlWebServer(t *testing.T) {
	curlService(t, "nginx", "", "80", "200")
}

func curlLoadBalancer(t *testing.T){
	curlServiceLB(t, "nginx", "", "80", "200")
}

func sshLoadBalancer(t *testing.T){
	err := ssh.CheckSshConnectionE(t, lbHost(t))  
	if err == nil {
		t.Fatalf("error in calling vcn: %s", err.Error())
	}

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

func checkCompute(t *testing.T) {
	// client
	config := common.CustomProfileConfigProvider("", "CzechEdu")
	c, _ := core.NewComputeClientWithConfigurationProvider(config)
	// c, _ := core.NewVirtualNetworkClientWithConfigurationProvider(common.DefaultConfigProvider())

	// request
	request := core.GetInstanceRequest{}
	instanceId := sanitizeInstanceId(t)
	request.InstanceId = &instanceId

	//t.Fatalf(instanceId)

	// response
	response, err := c.GetInstance(context.Background(), request)

	if err != nil {
		t.Fatalf("error in calling instance: %s", err.Error())
	}

	// assertions
	expected := "webServer0-default"
	actual := response.Instance.DisplayName

	if expected != *actual {
		t.Fatalf("wrong instance display name: expected %q, got %q", expected, *actual)
	}

	expected = "VM.Standard2.1"
	actual = response.Shape

	if expected != *actual {
		t.Fatalf("wrong shape: expected %q, got %q", expected, *actual)
	}
}

func checkLoadBalancer(t *testing.T) {
	// client
	config := common.CustomProfileConfigProvider("", "CzechEdu")
	c, _ := core.NewVirtualNetworkClientWithConfigurationProvider(config)

	// request
	request := core.GetSubnetRequest{}
	loadBalancerId := sanitizeLoadBalancerId(t)
	request.SubnetId = &loadBalancerId

	// response
	response, err := c.GetSubnet(context.Background(), request)

	if err != nil {
		t.Fatalf("error in calling vcn: %s", err.Error())
	}

	// assertions
	expected := "Loadbalancer Subnet-default"
	actual := response.Subnet.DisplayName

	if expected != *actual {
		t.Fatalf("wrong subnet display name: expected %q, got %q", expected, *actual)
	}

	expected = "10.0.200.0/28"
	actual = response.Subnet.CidrBlock

	if expected != *actual {
		t.Fatalf("wrong cidr block: expected %q, got %q", expected, *actual)
	}
	
	secListId := response.Subnet.SecurityListIds[0]
	secListRequest := core.GetSecurityListRequest{}
	secListRequest.SecurityListId = &secListId

	newresponse, newerr := c.GetSecurityList(context.Background(), secListRequest)

	if newerr != nil {
		t.Fatalf("error in calling seclist: %s", newerr.Error())
	}

	outexpected := 1
	outactual := len(newresponse.SecurityList.EgressSecurityRules)

	if outexpected != outactual {
		t.Fatalf("wrong number of egress rules: expected %q, got %q", outexpected, outactual)
	}

	inexpected := 2
	inactual := len(newresponse.SecurityList.IngressSecurityRules)

	if inexpected != inactual {
		t.Fatalf("wrong number of egress rules: expected %q, got %q", inexpected, inactual)
	}
}



func sanitizeInstanceId(t *testing.T) string {
	raw:= terraform.Output(t, options, "InstanceId")
	return strings.Split(raw, "\"")[1]
}

func sanitizeLoadBalancerId(t *testing.T) string {
	raw:= terraform.Output(t, options, "LoadBalancerId")
	return strings.Split(raw, "\"")[1]
}


func sanitizedVcnId(t *testing.T) string {
	raw := terraform.Output(t, options, "VcnID")
	return strings.Split(raw, "\"")[1]
}

// ~~~~~~~~~~~~~~~~ Helper functions ~~~~~~~~~~~~~~~~
func lbHost(t *testing.T) ssh.Host {
	lbIP := terraform.OutputList(t, options, "lb_ip")[0]
	return sshHost(t, lbIP)
}

func bastionHost(t *testing.T) ssh.Host {
	bastionIP := terraform.OutputList(t, options, "BastionPublicIP")[0]
	return sshHost(t, bastionIP)
}

func webHost(t *testing.T) ssh.Host {
	webIP := terraform.OutputList(t, options, "WebServerPrivateIPs")[0]
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
func curlServiceLB(t *testing.T, serviceName string, path string, port string, returnCode string) {
	//loadBHost := lbHost(t)
	host := loadBalancerIPs(t)
	url := fmt.Sprintf("http://%s:%s%s", host, port, path)
	curl := exec.Command("curl", url)
	_, err := curl.Output()
	if err != nil {
		t.Fatalf("error curl to %s", url)
		return
	}
}


func curl(host string, port string, path string) string {
	//return fmt.Sprintf("curl http://%s:%s%s", host, port, path)
	return fmt.Sprintf("curl -s -o /dev/null -w '%%{http_code}' http://%s:%s%s", host, port, path)
}


func webServerIPs(t *testing.T) []string {
	return terraform.OutputList(t, options, "WebServerPrivateIPs")
}

func loadBalancerIPs(t *testing.T) string {
	return terraform.OutputList(t, options, "lb_ip")[0]
}

func jumpSsh(t *testing.T, command string, expected string, retryAssert bool) string {
	bastionHost := bastionHost(t)
	webHost := webHost(t)
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
	jumpSsh(t, command, strconv.Itoa(expectedCount), true)
}
