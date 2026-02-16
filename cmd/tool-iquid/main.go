package main

import (
	"flag"
	"time"

	"github.com/DavidMANZI-093/tool-iquid/pkg/core"
	"github.com/DavidMANZI-093/tool-iquid/pkg/liquid"
	"github.com/DavidMANZI-093/tool-iquid/pkg/utils"
)

func main() {
	routerURL := flag.String("url", "http://192.168.1.254", "Router URL")
	username := flag.String("user", "", "Router Username")
	password := flag.String("pass", "", "Router Password")
	targetSSID := flag.String("ssid", "Target_SSID_Name", "Target SSID to watch")
	cooldown := flag.Duration("cooldown", 5*time.Minute, "Cooldown time after reboot")
	checkInterval := flag.Duration("interval", 30*time.Second, "Internet check interval")
	dryRun := flag.Bool("dry-run", false, "Perform checks and login but do not reboot")
	force := flag.Bool("force", false, "Force login check even if internet is online")
	configFile := flag.String("config", "", "Path to configuration file (default: ~/.tool-iquid/config.json or XDG)")

	flag.Parse()

	configPath := *configFile
	if configPath == "" {
		configPath = core.FindConfigFile()
	}

	cfg, err := core.LoadConfig(configPath)
	if err == nil {
		if *routerURL == "http://192.168.1.254" && cfg.RouterURL != "" {
			*routerURL = cfg.RouterURL
		}
		if *username == "" && cfg.Username != "" {
			*username = cfg.Username
		}
		if *password == "" && cfg.Password != "" {
			*password = cfg.Password
		}
		if *targetSSID == "Target_SSID_Name" && cfg.TargetSSID != "" {
			*targetSSID = cfg.TargetSSID
		}
		if *cooldown == 5*time.Minute && cfg.Cooldown != 0 {
			*cooldown = cfg.Cooldown
		}
		if *checkInterval == 30*time.Second && cfg.CheckInterval != 0 {
			*checkInterval = cfg.CheckInterval
		}
	}

	utils.LogInfo("Starting tool-iquid. Watching SSID: %s", *targetSSID)

	for {
		ssid, err := core.GetCurrentSSID()
		if err != nil {
			utils.LogWarn("Error getting SSID: %v", err)
			time.Sleep(5 * time.Second)
			continue
		} else if ssid != *targetSSID {
			utils.LogWarn("Not connected to target SSID (Current: %s). Waiting...", ssid)
			time.Sleep(*checkInterval)
			continue
		} else {
			utils.LogInfo("Connected to %s", ssid)
		}

		online := core.CheckConnectivity()
		if online {
			utils.LogSuccess("Internet is ONLINE")
		} else {
			utils.LogWarn("Internet is OFFLINE")
		}

		if !online || *force || *dryRun {
			if *force && online {
				utils.LogWarn("Force mode enabled. Initiating recovery check...")
			} else {
				utils.LogWarn("Initiating recovery sequence...")
			}

			cooldownVal := 5 * time.Minute
			if cfg != nil && cfg.Cooldown > 0 {
				cooldownVal = cfg.Cooldown
			}

			timeoutVal := 30 * time.Second
			if cfg != nil && cfg.Timeout > 0 {
				timeoutVal = cfg.Timeout
			}

			client := liquid.NewClient(*routerURL, *username, *password, timeoutVal)

			if *dryRun {
				utils.LogInfo("[Dry-Run] Would try to login and reboot now.")
				if err := client.Login(); err != nil {
					utils.LogError("[Dry-Run] Login check failed: %v", err)
				} else {
					utils.LogSuccess("[Dry-Run] Login check successful!")
				}
			} else {
				utils.LogInfo("Logging in...")
				if err := client.Login(); err != nil {
					utils.LogError("Login failed: %v", err)
					time.Sleep(*checkInterval)
					continue
				}

				utils.LogSuccess("Login successful. Attempting reboot...")
				if err := client.Reboot(); err != nil {
					utils.LogError("Reboot failed: %v", err)
				} else {
					utils.LogSuccess("Reboot triggered successfully! Waiting %d minutes for restart...", cooldownVal)
					time.Sleep(cooldownVal)
				}
			}
		}

		time.Sleep(*checkInterval)
	}
}
