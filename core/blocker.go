package suricata

import (
	"fmt"
	"log/slog" // ← DODANE
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
		slog.Error("Firewall block command failed", "ip", ip, "output", string(output), "error", err) // ← DODANE
		return fmt.Errorf("error during blocking IP %s: %v, output: %s", ip, err, string(output))
	}

	reloadCmd := exec.Command("sudo", "firewall-cmd", "--reload")
	if out, err := reloadCmd.CombinedOutput(); err != nil {
		slog.Error("Firewall reload failed", "output", string(out), "error", err) // ← DODANE
		return fmt.Errorf("error during reloading firewalld service: %v, (%s)", err, string(out))
	}

	slog.Info("IP blocked in firewall", "ip", ip) // ← DODANE
	return nil
}

func UnblockIP(ip string) error {
	cmd := exec.Command("sudo", "firewall-cmd",
		"--permanent",
		"--remove-rich-rule",
		fmt.Sprintf("rule family=ipv4 source address=%s drop", ip),
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		slog.Error("Firewall unblock command failed", "ip", ip, "output", string(output), "error", err) // ← DODANE
		return fmt.Errorf("error during unblocking IP %s: %v, output: %s", ip, err, string(output))
	}

	reloadCmd := exec.Command("sudo", "firewall-cmd", "--reload")
	if out, err := reloadCmd.CombinedOutput(); err != nil {
		slog.Error("Firewall reload failed", "output", string(out), "error", err) // ← DODANE
		return fmt.Errorf("error during reloading firewalld service: %v, (%s)", err, string(out))
	}

	slog.Info("IP unblocked in firewall", "ip", ip) // ← DODANE
	return nil
}
