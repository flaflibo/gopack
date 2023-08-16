package docker

import (
	"context"
	"fmt"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	sdkClient "github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

type Docker struct {
	dockerClient    *sdkClient.Client
	enabled         bool
	fluentdEnabled  bool
	fluentdHostname string
	fluentdConfig   string
	fluentdIp       string
	networkId       string
	networkSubnet   string
	networkGateway  string
}

func (d *Docker) Init(fluentdEnable bool, fluentdHostname, fluentdConfig, fluentdIp, networkId, networkSubnet string, networkGateway string, enabled bool) error {
	var err error
	d.dockerClient, err = sdkClient.NewClientWithOpts(sdkClient.FromEnv)
	if err != nil {
		return err
	}

	d.fluentdEnabled = fluentdEnable
	d.fluentdHostname = fluentdHostname
	d.fluentdConfig = fluentdConfig
	d.networkId = networkId
	d.networkSubnet = networkSubnet
	d.enabled = enabled
	d.networkGateway = networkGateway
	d.fluentdIp = fluentdIp

	return nil
}

func (d *Docker) Enabled() bool {
	return d.enabled
}

func (d *Docker) NetworkCreate(id, subnet string) (string, string, error) {
	ctx := context.Background()

	options := map[string]string{
		// "com.docker.network." + id + ".enable_icc":           "true",
		// "com.docker.network." + id + ".enable_ip_masquerade": "true",
		// "com.docker.network." + id + ".host_binding_ipv4":    "0.0.0.0",
		// "com.docker.network.driver.mtu":                      "1500",
	}

	netData, err := d.dockerClient.NetworkInspect(ctx, id, types.NetworkInspectOptions{})
	if err != nil {
		//Some error so network might not exist

		resp, err := d.dockerClient.NetworkCreate(ctx, id, types.NetworkCreate{
			Driver: "bridge",

			IPAM: &network.IPAM{
				Config: []network.IPAMConfig{
					{Subnet: subnet, Gateway: d.networkGateway},
				},
			},
			Options: options,
		})

		if err != nil {
			return "", "", err
		}

		return resp.ID, resp.Warning, nil
	}

	return netData.ID, "", nil
}

func (d *Docker) PullImage(imageName string) {
	ctx := context.Background()
	d.dockerClient.ImagePull(ctx, imageName, types.ImagePullOptions{})
}

func (d *Docker) StartFluentBit(containerName string) error {
	d.PullImage("fluent/fluent-bit:2.1.8")

	_, id := d.ContainerRunning(d.fluentdHostname)
	if id != "" {
		ctx := context.Background()
		d.dockerClient.ContainerRemove(ctx, id, types.ContainerRemoveOptions{Force: true})
	}

	cmd := []string{
		"-c", d.fluentdConfig,
	}

	_, _, err := d.ContainerCreateAndStart(
		"fluent/fluent-bit:2.1.8",
		"0:0",
		containerName,
		"always",
		d.fluentdIp,
		[]string{},
		[]string{d.fluentdConfig + ":" + d.fluentdConfig},
		nil,
		cmd,
	)

	return err
}

// checks if a container is running
// return bool --> true: container is running, false container is stopped
// return *string -->  "": container not found, id of the cointainer if it was found
func (d *Docker) ContainerRunning(containerName string) (bool, string) {
	ctx := context.Background()
	containerList, _ := d.dockerClient.ContainerList(ctx, types.ContainerListOptions{All: false})
	for _, container := range containerList {

		if container.Names[0] == fmt.Sprintf("/%s", containerName) {
			return true, container.ID
		}
	}

	//Check if its under the running servers --> there must be an easier way than this
	containerList, _ = d.dockerClient.ContainerList(ctx, types.ContainerListOptions{All: true})
	for _, container := range containerList {

		if container.Names[0] == fmt.Sprintf("/%s", containerName) {
			return false, container.ID
		}
	}

	return false, ""
}

// create a docker container and directly start it
// @params fluentBitHost: the docker hostname for the fluent bit container if available on the system.
//
//	setting this values will try to enable logging via fluentBit to the location of fluent
//	fluent-bit setup. fluentbit must be connected to the default network "bridge"
func (d *Docker) ContainerCreateAndStart(imageName, user, containerName, restart, ip string, ports, volumes, environment, commands []string) (string, []string, error) {
	ctx := context.Background()

	d.PullImage(imageName)

	//prepare ports
	exposedPorts := make(nat.PortSet)
	portBinding := nat.PortMap{}
	for _, port := range ports {
		tmp := strings.Split(port, ":")
		portBinding[nat.Port(tmp[0])] = []nat.PortBinding{
			{
				// HostIP:   "0.0.0.0",
				HostPort: tmp[1],
			},
		}
		port := nat.Port(tmp[0])
		exposedPorts[port] = struct{}{}
	}

	//prepare mounts
	mounts := []mount.Mount{}

	for _, volume := range volumes {
		tmp := strings.Split(volume, ":")
		m := mount.Mount{
			Type:   mount.TypeBind,
			Source: tmp[0],
			Target: tmp[1],
		}
		mounts = append(mounts, m)
	}

	config := container.Config{
		Tty:             false,
		AttachStdin:     true,
		AttachStdout:    true,
		Env:             environment,
		WorkingDir:      "/app", //every docker container that needs sources must put it under /app
		Image:           imageName,
		User:            user,
		ExposedPorts:    exposedPorts,
		NetworkDisabled: false,
		Cmd:             commands,
		Hostname:        containerName,
	}

	//
	// Enable fluent bit logging if fluent bit host is set.
	// Fluent bit contianer must be configured on the system separately
	// and running in the same docker network "bridge" (default network)
	//
	var logConfig container.LogConfig = container.LogConfig{}
	if d.fluentdEnabled && !strings.Contains(imageName, "fluent-bit") {
		//at this point fluent bit container is available and running so we can set the
		//docker log driver to use fluent bit
		logConfig = container.LogConfig{
			Type: "fluentd",
			Config: map[string]string{
				"fluentd-address": fmt.Sprintf("tcp://%s:24224", d.fluentdIp),
				"tag":             "system-agent-log",
			},
		}

	}

	hostConfig := container.HostConfig{
		Mounts:       mounts,
		AutoRemove:   false,
		PortBindings: portBinding,
		RestartPolicy: container.RestartPolicy{
			Name: restart,
		},
		LogConfig:   logConfig,
		NetworkMode: container.NetworkMode(d.networkId),
	}

	endPoint := network.EndpointSettings{
		Aliases:   []string{containerName},
		NetworkID: d.networkId,
		IPAMConfig: &network.EndpointIPAMConfig{
			IPv4Address: ip,
		},
	}

	endpointsConfig := make(map[string]*network.EndpointSettings)
	endpointsConfig[d.networkId] = &endPoint

	dockCont, err := d.dockerClient.ContainerCreate(ctx, &config, &hostConfig, &network.NetworkingConfig{
		EndpointsConfig: endpointsConfig,
	}, nil, containerName)
	if err != nil {
		return "", nil, err
	}

	err = d.dockerClient.ContainerStart(ctx, dockCont.ID, types.ContainerStartOptions{})
	if err != nil {
		return "", nil, err
	}

	return dockCont.ID, dockCont.Warnings, err
}
