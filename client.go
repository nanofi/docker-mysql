package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	docker "github.com/fsouza/go-dockerclient"
)

type DockerContainer struct {
}

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func getEndpoint() (string, error) {
	defaultEndpoint := "unix:///var/run/docker.sock"
	if os.Getenv("DOCKER_HOST") != "" {
		defaultEndpoint = os.Getenv("DOCKER_HOST")
	}
	if endpoint != "" {
		defaultEndpoint = endpoint
	}
	proto, host, err := parseHost(defaultEndpoint)
	if err != nil {
		return "", err
	}
	if proto == "unix" {
		exist, err := exists(host)
		if err != nil {
			return "", err
		}
		if !exist {
			return "", errors.New(host + " does not exist")
		}
	}
	return defaultEndpoint, nil
}

// based off of https://github.com/dotcloud/docker/blob/2a711d16e05b69328f2636f88f8eac035477f7e4/utils/utils.go
func parseHost(addr string) (string, string, error) {
	var (
		proto string
		host  string
		port  int
	)
	addr = strings.TrimSpace(addr)
	switch {
	case addr == "tcp://":
		return "", "", fmt.Errorf("Invalid bind address format: %s", addr)
	case strings.HasPrefix(addr, "unix://"):
		proto = "unix"
		addr = strings.TrimPrefix(addr, "unix://")
		if addr == "" {
			addr = "/var/run/docker.sock"
		}
	case strings.HasPrefix(addr, "tcp://"):
		proto = "tcp"
		addr = strings.TrimPrefix(addr, "tcp://")
	case strings.HasPrefix(addr, "fd://"):
		return "fd", addr, nil
	case addr == "":
		proto = "unix"
		addr = "/var/run/docker.sock"
	default:
		if strings.Contains(addr, "://") {
			return "", "", fmt.Errorf("Invalid bind address protocol: %s", addr)
		}
		proto = "tcp"
	}

	if proto != "unix" && strings.Contains(addr, ":") {
		hostParts := strings.Split(addr, ":")
		if len(hostParts) != 2 {
			return "", "", fmt.Errorf("Invalid bind address format: %s", addr)
		}
		if hostParts[0] != "" {
			host = hostParts[0]
		} else {
			host = "127.0.0.1"
		}

		if p, err := strconv.Atoi(hostParts[1]); err == nil && p != 0 {
			port = p
		} else {
			return "", "", fmt.Errorf("Invalid bind address format: %s", addr)
		}

	} else if proto == "tcp" && !strings.Contains(addr, ":") {
		return "", "", fmt.Errorf("Invalid bind address format: %s", addr)
	} else {
		host = addr
	}
	if proto == "unix" {
		return proto, host, nil

	}
	return proto, fmt.Sprintf("%s:%d", host, port), nil
}

func getContainers(client *docker.Client) ([]*RuntimeContainer, error) {
	apiContainers, err := client.ListContainers(docker.ListContainersOptions{
		All:  false,
		Size: false,
	})
	if err != nil {
		return nil, err
	}
	containers := []*RuntimeContainer{}
	for _, apiContainer := range apiContainers {
		container, err := client.InspectContainer(apiContainer.ID)
		if err != nil {
			log.Printf("error inspecting container: %s: %s\n", apiContainer.ID, err)
			continue
		}
		runtimeContainer := &RuntimeContainer{
			ID:      container.ID,
			Name:    strings.TrimLeft(container.Name, "/"),
			Address: container.NetworkSettings.IPAddress,
			Env:     make(map[string]string),
		}
		for _, entry := range container.Config.Env {
			parts := strings.Split(entry, "=")
			runtimeContainer.Env[parts[0]] = parts[1]
		}
		containers = append(containers, runtimeContainer)
	}
	return containers, nil
}
