# gonmmm

A Go interface for NetworkManager (`nmcli`) and ModemManager (`mmcli`) — manage network connections and GSM/LTE modems on Linux.

## Installation

```bash
go get github.com/sonnt85/gonmmm
```

> **Requirements:** `nmcli` (NetworkManager) and `mmcli` (ModemManager) must be installed and available in `$PATH`.

## Features

### NetworkManager (`NM*` functions)
- Check, create, delete, enable, disable, and restart connections
- Get and set connection field values
- Configure static IP or DHCP on a connection
- Create Wi-Fi connections with WPA-PSK
- List active devices sorted by connection type priority
- Reload NetworkManager configuration

### ModemManager (`MM*` functions)
- Detect the active GSM/LTE modem index
- Run raw `mmcli` commands
- Send AT commands to the modem
- Scan available cellular networks
- Query modem fields (operator name, signal strength, SIM state, SIM number)
- Reset the modem

## Usage

```go
package main

import (
    "fmt"

    "github.com/sonnt85/gonmmm"
)

func main() {
    // Check if a connection exists and bring it up
    if gonmmm.NMConIsExist("my-eth") {
        gonmmm.NMEnableCon("my-eth")
    }

    // Create a Wi-Fi connection
    gonmmm.NMCreateWifiConnection("home-wifi", "wlan0", "MySSID", "mypassword")

    // Configure a static IP
    gonmmm.NMSetStaticCon("my-eth", "192.168.1.100", "255.255.255.0", "192.168.1.1", "8.8.8.8")

    // Switch to DHCP
    gonmmm.NMActiveDhcpCon("my-eth")

    // Read modem info
    fmt.Println("Operator:", gonmmm.MMGetField("modem.3gpp.operator-name"))
    fmt.Println("Signal:", gonmmm.MMGetNetworkSignalStrength())

    // Send an AT command
    resp, _ := gonmmm.MMSendAtCommand("AT+CIMI")
    fmt.Println("IMSI:", resp)

    // Scan available networks
    networks, _ := gonmmm.MMGetListNetwork()
    for code, info := range networks {
        fmt.Printf("MCC/MNC: %s, Operator: %s, Tech: %s, Status: %s\n", code, info[0], info[1], info[2])
    }
}
```

## API

### NetworkManager

- `NMRunCommand(cmd string, timeouts ...time.Duration) (string, error)` — run an `nmcli` command
- `NMConIsExist(conname string) bool` — check if a connection exists
- `NMCheckConIsExist(conname string) error` — same, returns error
- `NMConIsActivated(conname string) bool` — check if a connection is active
- `NMDelCon(conname string) bool` — delete a connection
- `NMCleanupDuplicateCons(conname string) bool` — remove duplicate connections
- `NMEnableCon(conname string) error` — bring a connection up
- `NMDisableCon(conname string) error` — bring a connection down
- `NMRestartCon(conname string) error` — restart a connection
- `NMConGetField(conname, field string) string` — get a connection field value
- `NMConModField(conname, field, newval string, others ...string) error` — modify a connection field
- `NMConEditFieldIfChange(conname, field, newValue string) bool` — modify field only if value changed
- `NMCreateConnection(conname, ifacename, contype string, others ...string) error` — create a connection
- `NMCreateWifiConnection(conname, iface, ssid, password string, others ...string) error` — create a Wi-Fi connection
- `NMSetStaticCon(conname, ip, mask, gw, dns string) error` — configure static IPv4
- `NMActiveDhcpCon(conname string) error` — switch connection to DHCP
- `NMDisableDhcpCon(conname string) error` — disable DHCP (set a fallback static IP)
- `NMGetDevices() ([]string, error)` — list all network devices
- `NMGetInterfacesActivatedSortByType(types []string) ([]string, error)` — list active interfaces sorted by type priority
- `NMDevGetCon(ifacename string) string` — get the active connection name for a device
- `NMEnableDev/NMDisableDev(dev string)` — set device managed/unmanaged
- `NMReloadConfig()` — reload NetworkManager configuration
- `NMRestartMM()` — restart ModemManager service

### ModemManager

- `MMGetGsmIndex() string` — get the active modem index
- `MMRunCommand(cmd string, timeouts ...time.Duration) (string, error)` — run an `mmcli` command
- `MMSendAtCommand(cmd string, timeouts ...time.Duration) (string, error)` — send an AT command
- `MMGetListNetwork() (map[string][]string, error)` — scan available cellular networks
- `MMStatsGSM() string` — get full modem status output
- `MMGetCurrentOperator() (string, error)` — get current operator name
- `MMGetField(field string) string` — get a modem field value
- `MMGetFieldWithError(field string) (string, error)` — same, with error
- `MMGetSimState() (string, error)` — get SIM/modem state
- `MMGetSimNumber() string` — get the modem's own phone number
- `MMGetNetworkSignalStrength() string` — get signal quality percentage
- `MMResetGSM() (string, error)` — reset the modem

## License

MIT License - see [LICENSE](LICENSE) for details.
