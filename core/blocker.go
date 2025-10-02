package suricata

import (
	"fmt"
	"os/exec"
)

func BlockIP(ip string) error {
	cmd := exec.Command("sudo", "firewall-cmd",
		"--permanent",
		"--add-rich-rule",
		fmt.Sprintf("rule family=ipv4 source address=%s drop", ip),
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error during blocking IP %s: %v, output: %s", ip, err, string(output))
	}

	// Reload firewall to apply changes
	reloadCmd := exec.Command("sudo", "firewall-cmd", "--reload")
	if out, err := reloadCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("error during reloading firewalld service: %v, (%s)", err, string(out))
	}

	return nil
}
