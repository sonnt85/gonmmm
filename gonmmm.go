package gonmmm

import (
	"fmt"
	"strings"
	"time"

	"github.com/sonnt85/gosutils/gogrep"
	"github.com/sonnt85/gosutils/sexec"
	"github.com/sonnt85/gosutils/shellwords"
	"github.com/sonnt85/gosutils/sregexp"
	"github.com/sonnt85/gosutils/sutils"
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
	cmd := fmt.Sprintf(`con show %s`, conname)
	if _, err := NMRunCommand(cmd); err != nil {
		return false
	}
	return true
}

func NMDelCon(conname string) bool {
	cmd := fmt.Sprintf(`con show %s`, conname)
	if stdout, err := NMRunCommand(cmd); err != nil {
		return false
	} else {
		if strslides, err := gogrep.GrepStringLine(stdout, conname, -1, true); err == nil && len(strslides) > 1 {
			cmd = fmt.Sprintf(`con del %s`, conname)
			if _, err := NMRunCommand(cmd); err == nil {
				return false
			} else {
				return true
			}
		} else {
			return true
		}
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

func NMUpCon(conname string) error {
	if !NMConIsExist(conname) {
		return fmt.Errorf("Connection is not exist")
	}
	cmd := fmt.Sprintf(`con up %s`, conname)
	if _, err := NMRunCommand(cmd); err != nil {
		return fmt.Errorf("Can not up connection %s", err.Error())
		//		return err
	}
	return nil
}
