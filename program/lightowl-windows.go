package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/joho/godotenv"
)

var (
	outfile, _ = os.Open("C:\\Program Files\\lightowl\\lightowl.log") // update path for your needs
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
	lightowl_server := fmt.Sprintf("%s/api/v1/agents/config/%s", server, agent_id)

	req, err := http.NewRequest("GET", lightowl_server, nil)
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

func main() {
	err := godotenv.Load("C:\\Program Files\\lightowl\\.env")
	check(err)
	var LIGHTOWL_SERVER string = fmt.Sprintf("https://%s", os.Getenv("LIGHTOWL_SERVER"))
	var LIGHTOWL_AGENT_TOKEN string = os.Getenv("LIGHTOWL_AGENT_TOKEN")
	var LIGHTOWL_AGENT_ID string = os.Getenv("LIGHTOWL_AGENT_ID")

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
}
