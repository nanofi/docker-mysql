package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"sync"
	"time"

	docker "github.com/fsouza/go-dockerclient"
)

var (
	buildVersion string
	version      bool
	endpoint     string
	wg           sync.WaitGroup
)

type Event struct {
	ContainerID string `json:"id"`
	Status      string `json:"status"`
	Image       string `json:"image"`
}

type RuntimeContainer struct {
	ID      string
	Name    string
	Address string
	Env     map[string]string
}

func NewDockerClient(endpoint string) (*docker.Client, error) {
	return docker.NewClient(endpoint)
}

func update(client *docker.Client) {
	containers, err := getContainers(client)
	if err != nil {
		log.Printf("error listing containers: %s\n", err)
		return
	}
	dest, err := ioutil.TempFile("/tmp", "docker-mysql")
	defer func() {
		dest.Close()
		os.Remove(dest.Name())
	}()
	if err != nil {
		log.Fatalf("unable to create temp file: %s\n", err)
	}
	for _, container := range containers {
		if val, ok := container.Env["DB_NAME"]; ok {
			fmt.Fprintf(dest, "CREATE DATABASE IF NOT EXISTS %s ;\n", val)
		}
	}
	for _, container := range containers {
		user, okUser := container.Env["DB_USER"]
		pass, okPass := container.Env["DB_PASS"]
		db, okDb := container.Env["DB_NAME"]
		address := container.Address
		if okUser && okPass && okDb {
			fmt.Fprintf(dest, "GRANT ALL ON %s.* TO '%s'@'%s' identified by '%s' ;\n", db, user, address, pass)
		}
	}
	fmt.Fprintln(dest, "FLUSH PRIVILEGES ;")
	log.Printf("Configure database")
	cmdStr := fmt.Sprintf("mysql -u root -p$MYSQL_ROOT_PASSWORD < %s", dest.Name())
	cmd := exec.Command("/bin/sh", "-c", cmdStr)
	out, err := cmd.CombinedOutput()
	log.Println(string(out))
	if err != nil {
		log.Printf("Error running notify command: %s, %s\n", cmdStr, err)
		log.Print(string(out))
	}
}

func listen(client *docker.Client) {
	wg.Add(1)
	defer wg.Done()

	for {
		if client == nil {
			var err error
			endpoint, err := getEndpoint()
			if err != nil {
				log.Printf("Bad endpoint: %s", err)
				time.Sleep(10 * time.Second)
				continue
			}
			client, err = NewDockerClient(endpoint)
			if err != nil {
				log.Printf("Unable to connect to docker daemon: %s", err)
				time.Sleep(10 * time.Second)
				continue
			}
			update(client)
		}
		eventChan := make(chan *docker.APIEvents, 100)
		defer close(eventChan)

		watching := false
		for {
			if client == nil {
				break
			}
			err := client.Ping()
			if err != nil {
				log.Printf("Unable to ping docker daemon: %s", err)
				if watching {
					client.RemoveEventListener(eventChan)
					watching = false
					client = nil
				}
				time.Sleep(10 * time.Second)
				break
			}
			if !watching {
				err = client.AddEventListener(eventChan)
				if err != nil && err != docker.ErrListenerAlreadyExists {
					log.Printf("Error registering docker event listener: %s", err)
					time.Sleep(10 * time.Second)
					continue
				}
				watching = true
				log.Println("Watching docker events")
			}
			select {
			case event := <-eventChan:
				if event == nil {
					if watching {
						client.RemoveEventListener(eventChan)
						watching = false
						client = nil
					}
					break
				}
				if event.Status == "start" || event.Status == "stop" || event.Status == "die" {
					log.Printf("Received event %s for container %s", event.Status, event.ID[:12])
					update(client)
				}
			case <-time.After(10 * time.Second):
				// check for docker liveness
			}
		}
	}
}

func start(client *docker.Client) {
	update(client)
	listen(client)
}

func waitDbUp() {
	log.Printf("Waiting for starting database")
	cmdStr := "mysql -u root -p$MYSQL_ROOT_PASSWORD -e 'show databases;'"
	for {
		cmd := exec.Command("/bin/sh", "-c", cmdStr)
		_, err := cmd.CombinedOutput()
		if err == nil {
			break
		}
		log.Printf(".")
		time.Sleep(10 * time.Second)
	}
	log.Println("Alive database")
}

func initFlags() {
	flag.BoolVar(&version, "version", false, "show version")
	flag.Parse()
}

func main() {
	initFlags()

	if version {
		fmt.Println(buildVersion)
		return
	}

	endpoint, err := getEndpoint()
	if err != nil {
		log.Fatalf("Bad endpoint: %s", err)
	}

	client, err := NewDockerClient(endpoint)
	if err != nil {
		log.Fatalf("Unable to create docker client: %s", err)
	}

	waitDbUp()
	start(client)
	wg.Wait()
}
