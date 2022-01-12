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

var (
	outfile, _ = os.Open("C:\\Program Files\\lightowl\\lightowl.log")
	l          = log.New(outfile, "", 0)
)

func check(e error) {
	if e != nil {
		fmt.Println(e)
		panic(e)
	}
}

var LIGHTOWL_CONF_PATH string = "C:\\Program Files\\telegraf-1.21.1\\telegraf.d\\lightowl.conf"
var SSL_CA_PATH string = "C:\\Program Files\\lightowl\\ssl\\ca.pem"

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
	cmd := exec.Command("powershell", "(Get-service telegraf).status")
	output, _ := cmd.CombinedOutput()

	result := strings.TrimSpace(string(output))
	if strings.EqualFold(result, "Stopped") {
		fmt.Println("Telegraf isn't running. Starting it")
		cmd := exec.Command("C:\\Program Files\\telegraf-1.21.1\\telegraf.exe", "--service", "start")
		err := cmd.Run()
		fmt.Println(err)
		check(err)
	}
}

func send_installed_packages(server string, agent_token string, agent_id string) {
	fmt.Println("Fetching installed packages")
	type Dictionary map[string]interface{}
	var packages []Dictionary

	cmd := exec.Command(
		"powershell",
		"Get-ItemProperty",
		"HKLM:\\Software\\Wow6432Node\\Microsoft\\Windows\\CurrentVersion\\Uninstall\\*",
		"|",
		"Select-Object DisplayName, DisplayVersion, Publisher, InstallDate",
		"| ConvertTo-Csv -Delimiter '|' -NoTypeInformation",
	)

	output, err := cmd.CombinedOutput()
	check(err)

	result := strings.TrimSpace(string(output))
	installed_packages := strings.Split(string(result), "\n")
	r := regexp.MustCompile(`"(?P<software_name>.*)"\|"(?P<version>.*)"\|"(?P<vendor>.*)"\|.*`)

	for index, tmp := range installed_packages {
		if tmp == "|||" || index < 2 {
			continue
		}

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
			"name":    strings.Replace(result["software_name"], "\"", "", 0),
			"version": strings.Replace(result["version"], "\"", "", 0),
			"vendor":  strings.Replace(result["vendor"], "\"", "", 0),
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
	err := godotenv.Load("C:\\Program Files\\lightowl\\.env")
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
			cmd := exec.Command("C:\\Program Files\\telegraf-1.21.1\\telegraf.exe", "--service", "restart")
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
