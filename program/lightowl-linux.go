package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/joho/godotenv"
)

type software struct {
	name    string
	version string
}

var (
	outfile, _ = os.Open("/var/log/lightowl/lightowl.log")
	l          = log.New(outfile, "", 0)
)

func check(e error) {
	if e != nil {
		fmt.Println(e)
		panic(e)
	}
}

var LIGHTOWL_CONF_PATH string = "/etc/telegraf/telegraf.d/lightowl.conf"
var SSL_CA_PATH string = "/etc/ssl/lightowl/ca.pem"

func get_lightowl_config(server string, agent_token string, agent_id string) string {
	lightowl_url := fmt.Sprintf("%s/api/v1/agents/config/%s", server, agent_id)

	req, err := http.NewRequest("GET", lightowl_url, nil)
	check(err)

	req.Header.Set("api_key", agent_token)

	caCert, err := ioutil.ReadFile(SSL_CA_PATH)
	if err != nil {
		log.Fatalf("Error opening cert file %s, Error: %s", SSL_CA_PATH, err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	t := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: false, RootCAs: caCertPool},
	}

	client := http.Client{Transport: t}
	response, err := client.Do(req)
	check(err)

	if response.StatusCode != 200 {
		fmt.Println(response.StatusCode)
		panic("Error when requesting Lightowl Server")
	}

	defer response.Body.Close()

	result, err := ioutil.ReadAll(response.Body)
	check(err)

	config := string(result)
	config = strings.Replace(config, `\n`, "\n", -1)
	config = strings.Replace(config, "\"", "", -1)
	return config
}

func read_local_file() string {
	file, err := ioutil.ReadFile(LIGHTOWL_CONF_PATH)
	check(err)
	return string(file)
}

func check_telegraf_status() {
	cmd := exec.Command("/usr/bin/sudo", "/bin/systemctl", "check", "telegraf")
	_, err := cmd.CombinedOutput()

	if err != nil {
		fmt.Println("Telegraf isn't running. Starting it")
		cmd := exec.Command("/usr/bin/sudo", "/bin/systemctl", "start", "telegraf")
		err = cmd.Run()
		fmt.Println(err)
		check(err)
	}
}

func send_installed_packages(server string, agent_token string, agent_id string) {
	fmt.Println("Fetching installed packages")
	type Dictionary map[string]interface{}
	var packages []Dictionary

	cmd := exec.Command("/usr/bin/apt", "list", "--installed")
	res, err := cmd.CombinedOutput()
	check(err)

	installed_packages := strings.Split(string(res), "\n")
	r := regexp.MustCompile(`(?P<software_name>[\S]*)/.*? (?P<version>[\S]*)`)

	for _, tmp := range installed_packages {
		match := r.FindStringSubmatch(tmp)
		if len(match) == 0 {
			continue
		}

		result := make(map[string]string)
		for i, name := range r.SubexpNames() {
			if i != 0 && name != "" {
				result[name] = match[i]
			}
		}

		soft := Dictionary{
			"name":    result["software_name"],
			"version": result["version"],
		}

		packages = append(packages, soft)
	}

	lightowl_url := fmt.Sprintf("%s/api/v1/agents/packages/%s", server, agent_id)
	p, _ := json.Marshal(Dictionary{"softwares": packages})

	req, err := http.NewRequest("POST", lightowl_url, strings.NewReader(string(p)))
	check(err)

	req.Header.Set("api_key", agent_token)
	req.Header.Set("Content-Type", "application/json")

	caCert, err := ioutil.ReadFile(SSL_CA_PATH)
	if err != nil {
		log.Fatalf("Error opening cert file %s, Error: %s", SSL_CA_PATH, err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	t := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: false, RootCAs: caCertPool},
	}

	client := http.Client{Transport: t}
	response, err := client.Do(req)
	check(err)

	responseData, err := ioutil.ReadAll(response.Body)
	_ = responseData
	check(err)

	if response.StatusCode != 200 {
		fmt.Println(response.StatusCode)
		panic("Error when requesting Lightowl Server")
	}

	fmt.Println("Installed packages successfully updated")
}

func main() {
	err := godotenv.Load("/etc/lightowl/.env")
	check(err)
	var LIGHTOWL_SERVER string = fmt.Sprintf("https://%s", os.Getenv("LIGHTOWL_SERVER"))
	var LIGHTOWL_AGENT_TOKEN string = os.Getenv("LIGHTOWL_AGENT_TOKEN")
	var LIGHTOWL_AGENT_ID string = os.Getenv("LIGHTOWL_AGENT_ID")

	if len(os.Args) == 1 {
		local_file := read_local_file()
		remote_config := get_lightowl_config(LIGHTOWL_SERVER, LIGHTOWL_AGENT_TOKEN, LIGHTOWL_AGENT_ID)

		if strings.Compare(remote_config, local_file) == 1 || strings.Compare(local_file, remote_config) == 1 {
			fmt.Println("New configuration file from LightOwl. Write on disk")
			dataBytes := []byte(remote_config)
			err := ioutil.WriteFile(LIGHTOWL_CONF_PATH, dataBytes, 0)
			check(err)

			fmt.Println("Restarting Telegraf")
			cmd := exec.Command("/usr/bin/sudo", "/bin/systemctl", "restart", "telegraf")
			err = cmd.Run()
			check(err)
		} else {
			check_telegraf_status()
			fmt.Println("Configuration is valid.")
		}
	} else if os.Args[1] == "packages" {
		send_installed_packages(LIGHTOWL_SERVER, LIGHTOWL_AGENT_TOKEN, LIGHTOWL_AGENT_ID)
	}
}
