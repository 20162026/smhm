package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"log"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
)

func main() {
	var username = ""
	var password = ""
	var broker = ""
	var devname = ""
	var suppress = false
	var nontls = false
	var help = flag.Bool("help", false, "Show help")
	var timeintervalSeconds uint

	// TODO use a better library to parse command line arguments
	flag.StringVar(&username, "user", "", "mqtt username")
	flag.StringVar(&password, "pass", "", "mqtt password")
	flag.StringVar(&devname, "name", "device0", "device name used int db")
	flag.StringVar(&broker, "broker", "ssl://localhost:8883", "mqtt broker url must start with ssl://")
	flag.BoolVar(&suppress, "suppress", false, "suppress data logs")
	flag.BoolVar(&nontls, "notls", false, "use non tsl connection, broker url must start with tcp://")
	flag.UintVar(&timeintervalSeconds, "interval", 1, "data send time interval in seconds")
	flag.Parse()

	if timeintervalSeconds < 1 {
		timeintervalSeconds = 1
	}

	if *help {
		flag.Usage()
	} else {
		for {
			log.Println("Starting mqtt communication")
			log.Println("host: " + broker)
			log.Println("topic: " + "dev1/" + username + "/data/" + devname)
			opts := MQTT.NewClientOptions()
			opts.AddBroker(broker)
			opts.SetUsername(username)
			opts.SetPassword(password)
			if !nontls {
				opts.SetTLSConfig(&tls.Config{InsecureSkipVerify: true})
			}

			c := MQTT.NewClient(opts)
			if token := c.Connect(); token.Wait() && token.Error() != nil {
				log.Panic(token.Error())
				break
			}

			for {
				//get cpu info
				cpuInfo, err := cpu.Percent(1*time.Second, false)
				if err != nil {
					log.Println(err)
					time.Sleep(1 * time.Second)
					continue
				}

				//get memory info
				memInfo, err := mem.VirtualMemory()
				if err != nil {
					log.Println(err)
					time.Sleep(1 * time.Second)
					continue
				}

				//send cpu info
				type Mqtt_msg struct {
					Cpu_load float32
					Mem_load float32
				}

				t := Mqtt_msg{float32(cpuInfo[0]), float32(memInfo.UsedPercent)}
				m, err := json.Marshal(t)

				if !suppress {
					log.Printf("%v", string(m))
				}

				// TOPIC STRUCTURE dev1/{username}/data/{devname}
				// TODO add some kind of way to set topic using command line arguments
				token := c.Publish("dev1/"+username+"/data/"+devname, 0, false, string(m))
				if token.Wait() && token.Error() != nil {
					log.Println(token.Error())
					c.Disconnect(250)
					break
				}
			}

			time.Sleep(time.Duration(timeintervalSeconds) * time.Second)
		}
	}

}
