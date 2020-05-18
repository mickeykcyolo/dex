package model

import (
	"context"
	"errors"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
	"io"
	"io/ioutil"
	"os"
	"path"
	"runtime"
	"strings"
)

// ErrNothingToDo is returned whenever a request is made
// to update running at the latest image version.
var ErrNothingToDo = errors.New("nothing to do")

// Process represents a long-running process that is part
// of the system.
type Process struct {
	// Executable is the executable used to start the process.
	Executable string `json:"executable"`
	// Hostname is the hostname of the process host.
	Hostname string `json:"hostname"`
	// NumCPU is the number of logical CPU threads
	// available to the process.
	NumCPU int `json:"num_cpu"`
	// MemFree is the amount free memory in bytes.
	MemFree uint64 `json:"mem_free"`
	// MemUsed is the amount of used memory in bytes.
	MemUsed uint64 `json:"mem_used"`
	// MemTotal is the amount of total memory in bytes.
	MemTotal uint64 `json:"mem_total"`
	// DiskFree is the amount of free space, in bytes,
	// in the partition where the process executes.
	DiskFree uint64 `json:"disk_free"`
	// DiskUsed is the amount of the used space, in bytes,
	// in the partition where the process executes.
	DiskUsed uint64 `json:"disk_used"`
	// DiskTotal is the amount of total space, in bytes,
	// in the partition where the process executes.
	DiskTotal uint64 `json:"disk_total"`
	// Version represents the code-base version.
	Version Version `json:"version"`
	// ImageName is the name of the underlying container image, if any.
	ImageName string `json:"image_name"`
	// ImageHash is the hash of the underlying container image, if any.
	ImageHash string `json:"image_hash"`
	// LatestImageHash is the hash of the latest underlying
	// container image, if any.
	LatestImageHash string `json:"latest_image_hash"`
	// ContainerName is the name of the underlying
	// docker container, if any.
	ContainerName string `json:"container_name"`
	// ContainerId is the unique id of the underlying
	// docker container, if any.
	ContainerId string `json:"container_id"`
	// Networks is an array of network addresses associated
	// with the process.
	Networks []string `json:"networks"`
	// Stale maintains whether the process had been detected
	// as stale.
	Stale bool `json:"stale"`
}

// SelfProcess returns a Process representing the currently executing process.
func SelfProcess(ctx context.Context) (proc *Process) {

	// default to the background context:
	if ctx == nil {
		ctx = context.Background()
	}

	// create process object:
	proc = &Process{
		NumCPU:  runtime.NumCPU(),
		Version: LatestVersion,
	}

	// try append memory usage:
	if stat, err := mem.VirtualMemoryWithContext(ctx); err == nil {
		proc.MemFree = stat.Free
		proc.MemUsed = stat.Used
		proc.MemTotal = stat.Total
	}

	// try append disk usage:
	if stat, err := disk.UsageWithContext(ctx, "."); err == nil {
		proc.DiskFree = stat.Free
		proc.DiskUsed = stat.Used
		proc.DiskTotal = stat.Total
	}

	// try append network usage:
	if stat, err := net.InterfacesWithContext(ctx); err == nil {
		for _, stat := range stat {
			for _, addr := range stat.Addrs {
				proc.Networks = append(proc.Networks, addr.Addr)
			}
		}
	}

	// try append hostname:
	if hostname, err := os.Hostname(); err == nil {
		proc.Hostname = hostname
	}

	// try append executable:
	if executable, err := os.Executable(); err == nil {
		proc.Executable = path.Base(executable)
	}

	// try create docker client for information about the container:
	docker, err := client.NewClientWithOpts(client.WithAPIVersionNegotiation())

	if err != nil {
		return
	}

	// try to list containers:
	containers, err := docker.ContainerList(ctx, types.ContainerListOptions{})

	if err != nil {
		return
	}

	// loop through containers to find container with the same
	// executable name as this process:
	for _, container := range containers {
		// find the container for the current process, we do this
		// by finding a "running" container that's executable
		// contains the executable name.
		//
		// TODO: there MUST be a better way to go about this.
		if strings.Contains(container.Command, proc.Executable) && container.State == "running" {
			if container, err := docker.ContainerInspect(ctx, container.ID); err == nil {
				proc.ImageHash = container.Image
				proc.ImageName = container.Config.Image
				proc.ContainerId = container.ID
				proc.ContainerName = container.Name

				// will also populate LatestImageHash:
				proc.Stale, _ = proc.IsStale(ctx)
			}
			// no need to continue looping:
			break
		}
	}

	return
}

// IsStale returns true if the underlying image for proc is stale by checking its hash
// against the latest version in the docker host.
func (proc *Process) IsStale(ctx context.Context) (stale bool, err error) {

	// process might have already been detected as stale:
	if stale = proc.Stale; stale {
		return
	}

	var (
		// img is the reference image.
		img types.ImageInspect
		// resp is a placeholder for
		// docker pull response.
		resp io.ReadCloser
		// docker a the docker client.
		docker *client.Client
	)

	// default to a background context:
	if ctx == nil {
		ctx = context.Background()
	}

	// try create a docker client:
	if docker, err = client.NewClientWithOpts(client.WithAPIVersionNegotiation()); err != nil {
		return
	}

	// try to docker pull the latest image for proc from registry:
	if resp, err = docker.ImagePull(ctx, proc.ImageName, types.ImagePullOptions{}); err != nil {
		err = fmt.Errorf("failed image pull %q: %v", proc.ImageName, err)
		return
	}

	// close the response after were finished:
	defer resp.Close()

	// read response to end, otherwise the pull will be aborted:
	if _, err = io.Copy(ioutil.Discard, resp); err != nil {
		return
	}

	// inspect the image that's currently in the image store:
	if img, _, err = docker.ImageInspectWithRaw(ctx, proc.ImageName); err != nil {
		err = fmt.Errorf("failed inspect image %q: %v", proc.ImageName, err)
		return
	}

	// populate LatestImageHash:
	proc.LatestImageHash = img.ID

	// if the current image hash differs from the new image hash,
	// this process is considered stale:
	stale = proc.ImageHash != proc.LatestImageHash

	return
}

// Update runs a procedure to externally update proc if proc is, in fact, stale.
func (proc *Process) Update(ctx context.Context) (err error) {

	var (
		stale  bool
		body   container.ContainerCreateCreatedBody
		docker *client.Client
	)

	// default to background context:
	if ctx == nil {
		ctx = context.Background()
	}

	// try and find out if the image is stale:
	if stale, err = proc.IsStale(ctx); !stale {
		if err == nil {
			err = ErrNothingToDo
		}
		return
	}

	// create a docker client:
	if docker, err = client.NewClientWithOpts(client.WithAPIVersionNegotiation()); err != nil {
		return
	}

	const (
		// externalUpdaterImage is the image of the external updater.
		externalUpdaterImage = "containrrr/watchtower"
		// externalUpdaterCommand is the entry-point of the external updater.
		externalUpdaterCommand = "/watchtower"
		// externalUpdaterRunOnce is a flag to tell the external updater
		// to run once and then exit.
		externalUpdaterRunOnce = "--run-once"
		// externalUpdaterStopSignalLabel is the label name where
		// the external updater expects the stop signal value.
		externalUpdaterStopSignalLabel = "com.centurylinklabs.watchtower.stop-signal"
		// externalUpdaterStopSignal is the stop signal used
		// by the external updater.
		externalUpdaterStopSignal = "SIGHUP"
		// externalUpdaterName is the name that will be given
		// to the external updater container.
		externalUpdaterName = "cyolo-watchtower"
	)

	// kill external update container if already exists:
	_ = docker.ContainerKill(ctx, externalUpdaterName, "SIGTERM")

	// remove external update container if already exists:
	_ = docker.ContainerRemove(ctx, externalUpdaterName, types.ContainerRemoveOptions{Force: true})

	//  create container config to run external updater in:
	ctrCfg := &container.Config{
		Image: externalUpdaterImage,
		// tell the external updater to run once and only update the self container:
		Cmd: []string{externalUpdaterCommand, externalUpdaterRunOnce, proc.ContainerName},
		// use a label to instruct the external updater to use a custom termination signal:
		Labels: map[string]string{externalUpdaterStopSignalLabel: externalUpdaterStopSignal},
	}

	// bind the docker daemon socket to the external updater container:
	hostCfg := &container.HostConfig{
		Binds: []string{"/var/run/docker.sock:/var/run/docker.sock"},
	}

	// create the external update container:
	if body, err = docker.ContainerCreate(ctx, ctrCfg, hostCfg, nil, externalUpdaterName); err != nil {
		err = fmt.Errorf("failed to create external container: %v", err)
		return
	}

	// start the external update container:
	if err = docker.ContainerStart(ctx, body.ID, types.ContainerStartOptions{}); err != nil {
		err = fmt.Errorf("failed to start external container: %v", err)
		return
	}

	return
}

// Kill kills the process.
func (proc *Process) Kill() error {
	os.Exit(0)
	return nil
}
