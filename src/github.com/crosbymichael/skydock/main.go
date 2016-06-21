/*
   Multihost
   Multiple ports
*/

package main

import (
	"flag"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/crosbymichael/log"
	"github.com/crosbymichael/skydock/docker"
	"github.com/crosbymichael/skydock/utils"
	// influxdb "github.com/influxdb/influxdb/client"
	"github.com/skynetservices/skydns1/client"
	"github.com/skynetservices/skydns1/msg"
)

var (
	pathToSocket        string
	domain              string
	environment         string
	skydnsURL string
	skydnsContainerName string
	secret              string
	ttl                 int
	beat                int
	numberOfHandlers    int
	pluginFile          string

	skydns       Skydns
	dockerClient docker.Docker
	plugins      *pluginRuntime
	running      = make(map[string]struct{})
	runningLock  = sync.Mutex{}
)

func initFunc() {
	flag.StringVar(&pathToSocket, "s", "/var/run/docker.sock", "path to the docker unix socket")
	flag.StringVar(&skydnsURL, "skydns", "", "url to the skydns url")
	flag.StringVar(&skydnsContainerName, "name", "", "name of skydns container")
	flag.StringVar(&secret, "secret", "", "skydns secret")
	flag.StringVar(&domain, "domain", "", "same domain passed to skydns")
	flag.StringVar(&environment, "environment", "dev", "environment name where service is running")
	flag.IntVar(&ttl, "ttl", 60, "default ttl to use when registering a service")
	flag.IntVar(&beat, "beat", 0, "heartbeat interval")
	flag.IntVar(&numberOfHandlers, "workers", 3, "number of concurrent workers")
	flag.StringVar(&pluginFile, "plugins", "/plugins/rancher.js", "file containing javascript plugins (plugins.js)")

	flag.Parse()

}

func validateSettings() {
	if beat < 1 {
		beat = ttl - (ttl / 4)
	}

	if (skydnsURL != "") && (skydnsContainerName != "") {
		fatal(fmt.Errorf("specify 'name' or 'skydns', not both"))
	}

	if (skydnsURL == "") && (skydnsContainerName == "") {
		skydnsURL = "http://" + os.Getenv("SKYDNS_PORT_8080_TCP_ADDR") + ":8080"
	}

	if domain == "" {
		fatal(fmt.Errorf("Must specify your skydns domain"))
	}
}


func printCurrentSettings(){
	fmt.Println("Working with the fallowing parameters:")
	fmt.Printf("skydnsURL %s\n", skydnsURL)
	fmt.Printf("pathToSocket %s\n", pathToSocket)
	fmt.Printf("skydnsContainerName %s\n", skydnsContainerName)
	fmt.Printf("secret %s\n", secret)
	fmt.Printf("domain %s\n", domain)
	fmt.Printf("environment %s\n", environment)
	fmt.Printf("ttl %d\n", ttl)
	fmt.Printf("beat %d\n", beat)
	fmt.Printf("numberOfHandlers %d\n", numberOfHandlers)
	fmt.Printf("pluginFile %s\n", pluginFile)
}

func setupLogger() error {
	var (
		logger log.Logger
	)

	if host := os.Getenv("INFLUXDB_HOST"); host != "" {
		// config := &influxdb.ClientConfig{
		// 	Host:     host,
		// 	Database: os.Getenv("INFLUXDB_DATABASE"),
		// 	Username: os.Getenv("INFLUXDB_USER"),
		// 	Password: os.Getenv("INFLUXDB_PASSWORD"),
		// }
		//
		// logger, err = log.NewInfluxdbLogger(fmt.Sprintf("%s.%s", environment, domain), "skydock", config)
		// if err != nil {
		// 	return err
		// }
	} else {
		logger = log.NewStandardLevelLogger("skydock")
	}

	if err := log.SetLogger(logger); err != nil {
		return err
	}
	return nil
}

func heartbeat(uuid string, service *msg.Service) {
	runningLock.Lock()
	if _, exists := running[uuid]; exists {
		runningLock.Unlock()
		return
	}
	running[uuid] = struct{}{}
	runningLock.Unlock()

	defer func() {
		runningLock.Lock()
		delete(running, uuid)
		runningLock.Unlock()
	}()

	//var errorCount int
	for _ = range time.Tick(time.Duration(beat) * time.Second) {
		// checking if the service exists after updating
		_ , err2 := skydns.Get(uuid)
		if err2 != nil {
			log.Logf(log.ERROR, "The expected service is not registered in skydns. Trying to register service again: adding %s (%s) to skydns", uuid, service.Name)
			if err := skydns.Add(uuid, service); err != nil {
				log.Logf(log.ERROR, "Service (%s) %s  have not been addedd sucessfully. Error ", uuid, service.Name, err)
			}
		}

		err := updateService(uuid, ttl)
		if beat >=30 {
			if err != nil {
				log.Logf(log.INFO, "updating ttl for (%s) %s  fialed. Err: %s", uuid, service.Name, err)
			} else {
				log.Logf(log.INFO, "updating ttl for (%s) %s succeed", uuid, service.Name)
			}
		}
	}
}

// restoreContainers loads all running containers and inserts
// them into skydns when skydock starts
func restoreContainers() error {
	containers, err := dockerClient.FetchAllContainers()
	if err != nil {
		return err
	}

	var container *docker.Container
	for _, cnt := range containers {
		uuid := utils.Truncate(cnt.Id)
		if container, err = dockerClient.FetchContainer(uuid, cnt.Image); err != nil {
			if err != docker.ErrImageNotTagged {
				log.Logf(log.ERROR, "failed to fetch %s on restore: %s", cnt.Id, err)
			}
			continue
		}

		service, err := plugins.createService(container)
		if err != nil {
			// doing a fatal here because we cannot do much if the plugins
			// return an invalid service or error
			fatal(err)
		}
		if err := sendService(uuid, service); err != nil {
			log.Logf(log.ERROR, "failed to send %s to skydns on restore: %s", uuid, err)
		}
	}
	return nil
}

// sendService sends the uuid and service data to skydns
func sendService(uuid string, service *msg.Service) error {
	fmt.Printf("service.Name %s\n", service.Name)
	fmt.Printf("uuid %s\n", uuid)
	log.Logf(log.INFO, "adding %s (%s) to skydns", uuid, service.Name)
	err := skydns.Add(uuid, service);
	// ignore erros for conflicting uuids and start the heartbeat again
	if err == client.ErrConflictingUUID {
		log.Logf(log.INFO, "service already exists for %s. Resetting ttl.", uuid)
		updateService(uuid, ttl)
	}
	go heartbeat(uuid, service)
	return err
}

func removeService(uuid string) error {
	log.Logf(log.INFO, "removing %s from skydns", uuid)
	return skydns.Delete(uuid)
}

func addService(uuid, image string) error {
	container, err := dockerClient.FetchContainer(uuid, image)
	if err != nil {
		if err != docker.ErrImageNotTagged {
			return err
		}
		return nil
	}

	service, err := plugins.createService(container)
	if err != nil {
		// doing a fatal here because we cannot do much if the plugins
		// return an invalid service or error
		fatal(err)
	}

	if err := sendService(uuid, service); err != nil {
		return err
	}
	return nil
}

func updateService(uuid string, ttl int) error {
	return skydns.Update(uuid, uint32(ttl))
}

func eventHandler(c chan *docker.Event, group *sync.WaitGroup) {
	defer group.Done()

	for event := range c {
		log.Logf(log.DEBUG, "received event (%s) %s %s", event.Status, event.ContainerId, event.Image)
		uuid := utils.Truncate(event.ContainerId)

		switch event.Status {
		case "die", "stop", "kill":
			if err := removeService(uuid); err != nil {
				log.Logf(log.ERROR, "error removing %s from skydns: %s", uuid, err)
			}
		case "start", "restart":
			if err := addService(uuid, event.Image); err != nil {
				log.Logf(log.ERROR, "error adding %s to skydns: %s", uuid, err)
			}
		}
	}
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "%s\n", err)
	os.Exit(1)

}

func main() {
	initFunc()
	validateSettings()
	printCurrentSettings()
	if err := setupLogger(); err != nil {
		fatal(err)
	}

	var (
		err   error
		group = &sync.WaitGroup{}
	)

	plugins, err = newRuntime(pluginFile)
	if err != nil {
		fatal(err)
	}

	if dockerClient, err = docker.NewClient(pathToSocket); err != nil {
		log.Logf(log.FATAL, "error connecting to docker: %s", err)
		fatal(err)
	}

	if skydnsContainerName != "" {
		container, err := dockerClient.FetchContainer(skydnsContainerName, "")
		if err != nil {
			log.Logf(log.FATAL, "error retrieving skydns container '%s': %s", skydnsContainerName, err)
			fatal(err)
		}

		skydnsURL = "http://" + container.NetworkSettings.IpAddress + ":8080"
	}

	log.Logf(log.INFO, "skydns URL: %s", skydnsURL)

	if skydns, err = client.NewClient(skydnsURL, secret, domain, "172.17.42.1:53"); err != nil {
		log.Logf(log.FATAL, "error connecting to skydns: %s", err)
		fatal(err)
	}

	log.Logf(log.DEBUG, "starting restore of containers")
	if err := restoreContainers(); err != nil {
		log.Logf(log.FATAL, "error restoring containers: %s", err)
		fatal(err)
	}

	events := dockerClient.GetEvents()

	group.Add(numberOfHandlers)
	// Start event handlers
	for i := 0; i < numberOfHandlers; i++ {
		go eventHandler(events, group)
	}

	log.Logf(log.DEBUG, "starting main process")
	group.Wait()
	log.Logf(log.DEBUG, "stopping cleanly via EOF")
}
