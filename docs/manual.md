## tool-iquid

`tool-iquid` is a lightweight, daemonized Go utility designed to automate the reboot process of specific router models when internet connectivity is lost. It monitors network status, handles cryptographic authentication, and triggers a system restart on the target device to restore connectivity.

### Installation

#### Prerequisites
*   **Go**: Version 1.18 or higher is required to build the source.
*   **Systemd**: For daemon functionality (optional but recommended).
*   **Wireless Tools**: `libwifi` (via `github.com/mdlayher/wifi`) is used for SSID detection. Ensure your kernel supports Netlink (standard on modern Linux).

#### Building from Source
Clone the repository and build the binary:

```bash
git clone https://github.com/DavidMANZI-093/tool-iquid.git
cd tool-iquid
go build -o tool-iquid ./cmd/tool-iquid
```

To install globally:
```bash
sudo mv tool-iquid /usr/local/bin/
```

### Configuration

The tool uses a JSON configuration file. By default, it looks for `config.json` in the following locations (in order):
1.  Command-line argument (`-config`)
2.  `$XDG_CONFIG_HOME/tool-iquid/config.json`
3.  `$HOME/.config/tool-iquid/config.json`
4.  `$HOME/.tool-iquid/config.json`
5.  Current working directory.

#### Structure
Create a `config.json` file with the following structure:

```json
{
    "router_url": "http://192.168.1.254",
    "username": "admin",
    "password": "your_secure_password",
    "target_ssid": "Target_SSID_Name",
    "cooldown": 300000000000,
    "check_interval": 30000000000,
    "timeout": 30000000000
}
```

*   **router_url**: The base HTTP URL of the target router.
*   **target_ssid**: The specific SSID the tool must be connected to before attempting any actions. This prevents accidental reboots when connected to other networks.
*   **cooldown**: Duration (in nanoseconds) to wait after a reboot or startup. Default is 5 minutes (`300000000000`).
*   **check_interval**: Frequency of internet connectivity checks in nanoseconds. Default is 30 seconds (`30000000000`).

### Usage

Run the tool directly from the terminal or as a background service.

#### Command Line Flags

| Flag | Description | Default |
| :--- | :--- | :--- |
| `-config <path>` | Path to a specific JSON configuration file. | `config.json` |
| `-dry-run` | Perform login and network checks but **do not** trigger a reboot. Useful for testing credentials. | `false` |
| `-force` | Bypass internet connectivity checks and force a recovery sequence (login + reboot) immediately. | `false` |
| `-ssid <name>` | Override the target SSID to watch. | `Target_SSID_Name` |
| `-url <url>` | Override the router URL. | `http://192.168.1.254` |
| `-user <name>` | Override the username. | *(empty)* |
| `-pass <pass>` | Override the password. | *(empty)* |

#### Example
Perform a dry-run to verify authentication:
```bash
./tool-iquid -config config.json -dry-run
```

### Systemd Integration

To run `tool-iquid` as a system service, create a unit file.

1.  **Create Service File**: `/etc/systemd/system/tool-iquid.service`

    ```ini
    [Unit]
    Description=Automated Router Reboot Tool
    After=network-online.target
    Wants=network-online.target

    [Service]
    Type=simple
    User=your_username
    ExecStart=/usr/local/bin/tool-iquid -config /home/your_username/.config/tool-iquid/config.json
    Restart=on-failure
    RestartSec=5s

    [Install]
    WantedBy=multi-user.target
    ```

    *Note: Replace `your_username` and paths with appropriate values.*

2.  **Enable and Start**:
    ```bash
    sudo systemctl daemon-reload
    sudo systemctl enable --now tool-iquid
    ```

### Troubleshooting

#### Logs
The tool outputs color-coded logs to `stdout`.
*   **Cyan**: Informational messages.
*   **Green**: Successful operations (login, reboot trigger).
*   **Yellow**: Warnings (offline status, retries).
*   **Red**: Critical errors.

If running via systemd, view logs with `journalctl`:
```bash
journalctl -u tool-iquid -f
```

#### Common Issues
*   **"No active Wi-Fi connection found"**: Ensure the host has a wireless interface and is connected. The tool relies on Netlink/`libwifi` to query interface properties.
*   **"Login failed"**: Verify credentials in `config.json`. Use `-dry-run` to test.
*   **"Reboot failed"**: Ensure the router is reachable. Check firewall rules or IP address changes.

### Files
*   `/usr/local/bin/tool-iquid`: Main binary.
*   `~/.config/tool-iquid/config.json`: User configuration.
*   `/etc/systemd/system/tool-iquid.service`: Systemd unit file.
