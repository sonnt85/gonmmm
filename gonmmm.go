package gonmmm

import (
	"fmt"
	"net"
	"sort"
	"strings"
	"time"

	"github.com/sonnt85/gosutils/gogrep"
	"github.com/sonnt85/gosutils/sexec"
	"github.com/sonnt85/gosutils/shellwords"
	"github.com/sonnt85/gosutils/sregexp"
	"github.com/sonnt85/gosutils/sutils"
	"github.com/sonnt85/gosystem"
)

func MMGetGsmIndex() string {
	cmd := `index=$(mmcli -L | grep -oPe 'org[^\s]+' | grep -Poe '[0-9]+$')
	[[ $index ]] &&  echo -n "${index}";
	`
	if stdout, _, err := sexec.ExecCommandShell(cmd, time.Second*1); err != nil {
		fmt.Println("Can not get GsmDevice")
	} else {
		return sutils.StringTrimLeftRightNewlineSpace(string(stdout))
	}
	return ""
}

func NMConIsExist(conname string) bool {
	cmd := fmt.Sprintf(`-g connection.id con show %s`, conname)
	_, err := NMRunCommand(cmd)
	return err == nil
}

func NMCheckConIsExis(conname string) error {
	cmd := fmt.Sprintf(`-g connection.id con show %s`, conname)
	_, err := NMRunCommand(cmd)
	return err
}

func NMCleanupDuplicateCons(conname string) bool {
	cmd := fmt.Sprintf(`set -e; conname="%s"; x=$(nmcli -g connection.uuid  c s ${conname} | wc -l); ((x >=2)) && nmcli c delete ${conname}`, conname)
	if _, _, err := sexec.ExecCommandShell(cmd, time.Second*10); err == nil {
		return true
	}
	return false
}

func NMConIsActivated(conname string) bool {
	//	cmd := fmt.Sprintf(`-f GENERAL.STATE connection show %s`, conname)
	cmd := `-s -g NAME con show --active`

	if stdout, err := NMRunCommand(cmd); err != nil {
		return false
	} else {
		if ret, _ := gogrep.GrepStringLine(stdout, "^"+conname+"$", 1); len(ret) != 0 {
			return true
		}
	}
	return false
}

func NMDelCon(conname string) bool {
	if NMConIsExist(conname) {
		if _, err := NMRunCommand(fmt.Sprintf(`con del %s`, conname)); err == nil {
			return true
		} else {
			return false
		}
	} else {
		return true
	}
}

func MMRunCommand(cmd string, timeouts ...time.Duration) (stdout string, err error) {
	timeout := time.Second * 60
	if len(timeouts) != 0 {
		timeout = timeouts[0]
	}
	index := MMGetGsmIndex()
	if len(index) == 0 {
		return "", fmt.Errorf("Cannot found gsm")
	}
	if !strings.Contains(cmd, "--command=") {
		//		cmd = shellwords.Join(cmd)
	}
	mmtimeoutsecs := int(timeout / time.Second)
	cmd = fmt.Sprintf(`mmcli --timeout %d -m %s %s`, mmtimeoutsecs, index, cmd)
	//	log.Info("MM command: ", cmd)

	stdoutb, stderrb, err1 := sexec.ExecCommandShell(cmd, timeout)
	if err1 != nil {
		//		log.Errorf("MMCMD ERROR: [%s]", cmd)
		err1 = fmt.Errorf("%s", string(stderrb))
	}
	return string(stdoutb), err1
}

func MMSendAtCommand(cmd string, timeouts ...time.Duration) (stdout string, err error) {
	//	return MMRunCommand(fmt.Sprintf(`--command=%s`, shellwords.Join(cmd)))
	cmd = fmt.Sprintf(`--command=%s`, shellwords.Join(cmd))
	//	log.Info("ATcommand to send: ", cmd)

	stdout, err = MMRunCommand(cmd, timeouts...)
	if err != nil {
		return
	}

	if retregex := sregexp.New(`(?s)response:\s+'(.+)'$`).FindStringSubmatch(stdout); len(retregex) != 0 {
		return sutils.StringTrimLeftRightNewlineSpace(retregex[1]), nil
	}
	return
}

func MMGetListNetwork() (retstr map[string][]string, err error) {
	//Found 4 networks:
	//21404 - Yoigo (umts, available)
	//21407 - Movistar (umts, current)
	//21401 - vodafone ES (umts, forbidden)
	//21403 - Orange (umts, forbidden)
	retstr = make(map[string][]string)
	mmretstr := ""
	if mmretstr, err = MMRunCommand("--3gpp-scan", time.Minute*5); err == nil {
		for _, v := range sutils.String2lines(mmretstr) {
			//				fmt.Println(v)
			if ret := sregexp.New(`([0-9]+)\s+-\s+([^\s]+)\s+\(([^,]+),\s+([^\s,\)]+)`).FindStringSubmatch(v); len(ret) != 0 {
				retstr[ret[1]] = []string{ret[2], ret[3], ret[4]}
			}
		}
	} else {
		return
	}

	return
}

func MMStatsGSM() string {
	cmd := `index=$(mmcli -L | grep -oPe 'org[^\s]+' | grep -Poe '[0-9]+$')
	[[ $index ]] && mmcli -m ${index}
	`
	if stdout, _, err := sexec.ExecCommandShell(cmd, time.Second*1); err != nil {
		fmt.Println("Can not stats USB LTE")
	} else {
		return string(stdout)
	}
	return ""
}

func MMGetCurrentOperator() (retstr string, err error) {
	return MMGetFieldWithError("modem.3gpp.operator-name")
}

func MMGetFieldWithError(field string) (retstr string, err error) {
	retstr = ""
	err = nil
	if keyvalues, err1 := MMRunCommand("--output-keyvalue"); err1 == nil {
		if oname := sregexp.New(fmt.Sprintf(`%s\s+:\s+(.+)`, field)).FindStringSubmatch(keyvalues); len(oname) != 0 {
			// retstr = strings.TrimRight(oname[1], "\n")
			retstr = oname[1]
		}
	} else {
		err = err1
	}
	return retstr, err
}

func MMGetField(field string) (retstr string) {
	retstr, _ = MMGetFieldWithError(field)
	return retstr
}

func MMGetSimState() (retstr string, err error) {
	return MMGetFieldWithError(`modem.generic.state`)
}

func MMGetSimNumber() string {
	return MMGetField(`modem.generic.own-numbers.value\[1\]`)
}

func MMGetNetworkSignalStrength() string {
	return MMGetField(`modem.generic.signal-quality.value`)
}

func NMRunCommand(cmd string, timeouts ...time.Duration) (stdout string, err error) {
	timeout := time.Second * 20
	if len(timeouts) != 0 {
		timeout = timeouts[0]
	}
	nmtimeoutsecs := int(timeout / time.Second)

	//	cmd = shellwords.Join(cmd)
	cmd = fmt.Sprintf(`nmcli -w %d %s`, nmtimeoutsecs, cmd)
	//	log.Info("nmcli comman: ", cmd)
	stdoutb, stderrb, err1 := sexec.ExecCommandShell(cmd, timeout)
	if err1 != nil {
		//		log.Errorf("NMCMD ERROR: [%s]", cmd)
		err1 = fmt.Errorf("%s", string(stderrb))
	}
	return string(stdoutb), err1
}

func MMResetGSM() (stdout string, err error) {
	return MMRunCommand("-r")
}

func NMConGetField(conname, field string) string {
	if f, err := NMRunCommand(fmt.Sprintf("-s -g %s connection show %s", field, conname)); err == nil {
		// if len(f) != 0 {
		// 	lines := strings.Split(f, "\n")
		// 	return lines[0]
		// }
		return strings.TrimRight(f, "\n")
	}

	return ""
}

func NMConEditFieldIfChange(conname, field, newValue string) bool {
	if NMConGetField(conname, field) != newValue {
		if nil != NMConModField(conname, field, newValue) {
			return false
		}
	}
	return true
}

func NMConFieldIsMatch(conname, field, compareValue string) bool {
	return NMConGetField(conname, field) == compareValue
}

func NMConModField(conname, field, newval string, others ...string) error {
	other := ""
	if len(others) != 0 {
		other = fmt.Sprintf("%v", others)
		other = strings.Trim(other, "[]")
	}
	if _, err := NMRunCommand(fmt.Sprintf(`connection modify %s %s "%s" %s`, conname, field, newval, other)); err == nil {
		if NMConFieldIsMatch(conname, field, newval) {
			return nil
		} else {
			return fmt.Errorf("cannot set value %s for %s->%s", newval, conname, field)
		}
	} else {
		return err
	}
}

func NMEnableCon(conname string) error {
	if !NMConIsExist(conname) {
		return fmt.Errorf("Connection is not exist")
	}
	if !NMConIsActivated(conname) {
		cmd := fmt.Sprintf(`con up %s`, conname)
		if _, err := NMRunCommand(cmd); err != nil {
			return fmt.Errorf("Can not up connection %s", err.Error())
		}
	}
	return nil
}

func NMRestartCon(conname string) error {
	if !NMConIsExist(conname) {
		return fmt.Errorf("connection is not exist")
	}
	if NMConIsActivated(conname) {
		cmd := fmt.Sprintf(`con down %s`, conname)
		if _, err := NMRunCommand(cmd); err != nil {
			return fmt.Errorf("can not down connection %s", err.Error())
		}
	}
	cmd := fmt.Sprintf(`con up %s`, conname)
	if _, err := NMRunCommand(cmd); err != nil {
		return fmt.Errorf("can not up connection %s", err.Error())
	}
	return nil
}

func NMDisableCon(conname string) error {
	if !NMConIsExist(conname) {
		return nil
	}
	if NMConIsActivated(conname) {
		cmd := fmt.Sprintf(`con down %s`, conname)
		if _, err := NMRunCommand(cmd); err != nil {
			return fmt.Errorf("Can not down connection %s", err.Error())
		}
	}
	return nil
}

func NMUpCon(conname string) error {
	return NMEnableCon(conname)
}

func NMCreateConnection(conname, ifacename, contype string, others ...string) error {
	other := ""
	if len(others) != 0 {
		other = fmt.Sprintf("%v", others)
		other = strings.Trim(other, "[]")
	}
	//connection.autoconnect
	if NMConIsExist(conname) {
		oldIface := NMConGetField(conname, "connection.interface-name")
		oldContype := NMConGetField(conname, "connection.type")

		if oldIface != ifacename {
			if err := NMConModField(conname, "connection.interface-name", ifacename); err != nil {
				return fmt.Errorf(`%s\n"%s"->"%s"`, err.Error(), oldIface, ifacename)
			}
		}

		if strings.Contains(oldContype, "wireless") {
			oldContype = "wifi"
		}

		if !strings.Contains(oldContype, contype) {
			if err := NMConModField(conname, "connection.type", contype); err != nil {
				return err
			}
		}
		return nil
	}

	cmd := fmt.Sprintf(`connection add type %s ifname "%s" con-name %s %s`, contype, ifacename, conname, other)
	_, err := NMRunCommand(cmd)
	return err
}

func NMCreateWifiConnection(conname, iface, ssid, password string, others ...string) error {
	others = append(others, "ssid", ssid, "wifi-sec.key-mgmt", "wpa-psk", "wifi-sec.auth-alg", "open", "802-11-wireless-security.psk", password)
	return NMCreateConnection(conname, iface, "wifi", others...)
	// return NMCreateConnection(conname, iface, "wifi", "ssid", ssid, "wifi-sec.key-mgmt", "wpa-psk", "wifi-sec.auth-alg", "open", "802-11-wireless-security.psk")
}

func NMDisableDev(dev string) {
	NMRunCommand(fmt.Sprintf("device set %s managed no", dev))
}

func NMEnableDev(dev string) {
	NMRunCommand(fmt.Sprintf("device set %s  managed yes", dev))
}

func NMManagedDev(dev string) {
	NMRunCommand(fmt.Sprintf("device set %s managed yes", dev))
}

func NMUnManagedDev(dev string) {
	NMRunCommand(fmt.Sprintf("device set %s  managed no", dev))
}

func NMRestartMM() {
	gosystem.RestartApp("ModemManager")
}

func NMDevGetCon(ifacename string) string {
	if con, err := NMRunCommand(fmt.Sprintf("-s -g GENERAL.CONNECTION device show  %s", ifacename)); err == nil {
		return sutils.StringTrimLeftRightNewlineSpace(con)
	} else {
		return ""
	}
}

func NMSetStaticCon(connanme, ipstr, maskstr, gwipstr, dns string) error {
	if (len(ipstr) != 0 || len(maskstr) != 0) && NMConIsExist(connanme) {
		// cidrstr := ""
		cirdbits := 0
		if ip := net.ParseIP(ipstr); ip != nil {
			if len(maskstr) == 0 || maskstr == "0.0.0.0" {
				maskstr = net.IP(ip.DefaultMask()).String()
			}
		} else {
			return fmt.Errorf("ip %s is wrong", maskstr)
		}

		if mask := net.ParseIP(maskstr); mask == nil {
			return fmt.Errorf("mask %s is wrong", maskstr)
		} else {
			ip := net.ParseIP(ipstr)
			ipmask := net.IPMask(mask.To4())

			if len(gwipstr) == 0 {
				ipgw := ip.Mask(ipmask)
				ipgw[len(ipgw)-1] = ipgw[len(ipgw)-1] | 1
				gwipstr = net.IP(ipgw).String()
			}

			cirdbits, _ = ipmask.Size()
			// cidrstr = fmt.Sprintf("%s/%d", ip.Mask(ipmask).String(), cirdbits)
			ipstr = fmt.Sprintf("%s/%d", ipstr, cirdbits)
		}

		changed := false
		// cmd2run := fmt.Sprintf("ip a add %s/%s dev %s brd +", ipstr, maskstr, ifi)
		if !NMConFieldIsMatch(connanme, "ipv4.addresses", ipstr) {
			changed = true
			if err := NMConModField(connanme, "ipv4.addresses", ipstr); err != nil {
				return err
			}
		}

		if !NMConFieldIsMatch(connanme, "ipv4.gateway", gwipstr) {
			changed = true
			if err := NMConModField(connanme, "ipv4.gateway", gwipstr); err != nil {
				return err
			}
		}
		if len(dns) != 0 {
			changed = true
			if !NMConFieldIsMatch(connanme, "ipv4.dns", dns) {
				if err := NMConModField(connanme, "ipv4.dns", dns); err != nil {
					return err
				}
			}
			// dns = "1.1.1.1"
		}
		if !NMConFieldIsMatch(connanme, "ipv4.method", "manual") {
			changed = true
			if err := NMConModField(connanme, "ipv4.method", "manual"); err != nil {
				return err
			}
		}
		if changed {
			return NMRestartCon(connanme)
		}
		return nil

	} else {
		return fmt.Errorf("can not confiture ip with zero ip and mask for con %s", connanme)
	}
}

func NMActiveDhcpCon(connanme string) error {
	if NMConIsExist(connanme) {
		changed := false
		if !NMConFieldIsMatch(connanme, "ipv4.method", "auto") {
			changed = true
			if err := NMConModField(connanme, "ipv4.method", "auto"); err != nil {
				return err
			}
		}

		if !NMConFieldIsMatch(connanme, "ipv4.addresses", "") {
			changed = true
			if err := NMConModField(connanme, "ipv4.addresses", "", `ipv4.gateway ""`); err != nil {
				return err
			}
		}

		if changed {
			return NMRestartCon(connanme)
		} else {
			return nil
		}
	} else {
		return fmt.Errorf("missing connection %s", connanme)
	}
}

func NMDisableDhcpCon(connanme string) error {
	if NMConIsExist(connanme) {
		changed := false
		if !NMConFieldIsMatch(connanme, "ipv4.addresses", "192.168.168.168/24") {
			changed = true
			if err := NMConModField(connanme, "ipv4.addresses", "192.168.168.168/24"); err != nil {
				return err
			}
		}

		if !NMConFieldIsMatch(connanme, "ipv4.method", "manual") {
			changed = true
			if err := NMConModField(connanme, "ipv4.method", "manual"); err != nil {
				return err
			}
		}
		if changed {
			return NMRestartCon(connanme)
		} else {
			return nil
		}
	} else {
		return fmt.Errorf("missing connection %s", connanme)
	}
}

func NMGetDevices() (devices []string, err error) {
	if devices1, err1 := NMRunCommand("-s -g GENERAL.DEVICE device show"); err1 == nil {
		devices = strings.Split(devices1, "\n")
	} else {
		err = err1
	}
	return devices, err
}

func NMGetInterfacesActivedSortByType(types []string) (interfaces []string, err error) {
	var devices []string
	if devices1, err1 := NMRunCommand("-s -g DEVICE,TYPE con show -a"); err1 == nil {
		devices = strings.Split(devices1, "\n")

		sort.Slice(devices, func(i, j int) bool {
			devicesi := strings.Split(devices[i], ":")
			devicesj := strings.Split(devices[j], ":")
			i1 := 0
			j1 := 0
			cnt := 0
			for k, v := range types {
				if devicesi[1] == v {
					i1 = k
				}
				if devicesj[1] == v {
					j1 = k
				}
				if cnt == 2 {
					break
				}
			}
			if i1 < j1 {
				return true
			} else {
				return false
			}
		})
		for _, v := range devices {
			interfaces = append(interfaces, v)
		}
		return interfaces, nil
	} else {
		err = err1
	}
	return interfaces, err
}

func NMDeviceIsexist(conname string) bool {
	if devices, err := NMGetDevices(); err == nil {
		iname := NMConGetField(conname, "connection.interface-name")
		if len(iname) != 0 {
			return false
		}
		for _, v := range devices {
			if v == iname {
				return true
			}
		}
	}
	return false
}

func NMDelConIfDeviceIsNotexist(conname string) (ret bool) {
	if devices, err := NMGetDevices(); err == nil {
		iname := NMConGetField(conname, "connection.interface-name")
		if len(iname) != 0 {
			return false
		}
		for _, v := range devices {
			if v == iname {
				NMDelCon(conname)
				return true
			}
		}
	}
	return false
}

func NMReloadConfig() {
	NMRunCommand("reload")
}
