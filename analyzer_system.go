// analyzer_system.go
package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"sssd-inspector/constants"
)

// Global regex for parsing systemd exit codes
var reStatus = regexp.MustCompile(`status=(\d+)`)

func analyzeBasicHealth(dirPath string, report *ReportData) {
	scanFiles(dirPath, []string{"basic-health-check.txt", "basic-environment.txt"}, func(line string) {
		if strings.HasPrefix(line, "SR#:") {
			report.SupportCaseID = strings.TrimSpace(strings.TrimPrefix(line, "SR#:"))
		}
	})
}

func analyzeOSAndHardware(dirPath string, report *ReportData) {
	scanFiles(dirPath, []string{"basic-environment.txt"}, func(line string) {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Linux") && strings.Contains(line, "SMP") {
			report.KernelVersion = line
		}
		if strings.HasPrefix(line, "PRETTY_NAME=") {
			report.SLESRlease = strings.Trim(strings.Split(line, "=")[1], "\"")
		}
		if strings.HasPrefix(line, "Manufacturer:") {
			report.HardwareManufacturer = strings.TrimSpace(strings.TrimPrefix(line, "Manufacturer:"))
		}
		if strings.HasPrefix(line, "Hardware:") {
			report.HardwareModel = strings.TrimSpace(strings.TrimPrefix(line, "Hardware:"))
		}
		if strings.HasPrefix(line, "Hypervisor:") {
			report.Hypervisor = strings.TrimSpace(strings.TrimPrefix(line, "Hypervisor:"))
		}
		if strings.HasPrefix(line, "Identity:") {
			report.VirtualIdentity = strings.TrimSpace(strings.TrimPrefix(line, "Identity:"))
		}
	})
}

func analyzeHostnameAndFQDN(dirPath string, report *ReportData) {
	scanFiles(dirPath, []string{"basic-environment.txt"}, func(line string) {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Hostname:") {
			hostname := strings.TrimSpace(strings.TrimPrefix(line, "Hostname:"))
			if !strings.Contains(hostname, ".") {
				report.Problems = append(report.Problems, "[NETWORK] System is using a short hostname instead of an FQDN. Active Directory heavily relies on fully qualified domain names.")
			}
		}
	})
}

func analyzeSCC(dirPath string, report *ReportData) {
	sccFound := false
	scanFiles(dirPath, []string{"updates.txt"}, func(line string) {
		if strings.Contains(line, "Registered") || strings.Contains(line, "Status: Active") {
			report.SCCStatus = "Registered"
			sccFound = true
		}
	})
	if !sccFound {
		report.SCCStatus = "Not Registered or Unknown"
	}
}

func analyzePerformance(dirPath string, report *ReportData) {
	if anyFileContains(dirPath, []string{"memory.txt"}, "vm.dirty_bytes = 0") {
		report.Problems = append(report.Problems, "[PERFORMANCE] vm.dirty_bytes is set to 0. This disables limits on dirty data, leading to uncontrolled accumulation and severe I/O bottlenecks that can block SSSD.")
	}

	scanFiles(dirPath, []string{"sar.txt"}, func(line string) {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "all") && !strings.Contains(line, "CPU") {
			fields := strings.Fields(line)
			if len(fields) >= 6 {
				iowaitIndex := len(fields) - 3
				if iowait, err := strconv.ParseFloat(fields[iowaitIndex], 64); err == nil && iowait > 20.0 {
					report.Warnings = append(report.Warnings, fmt.Sprintf("[PERFORMANCE] High CPU %%iowait detected (>%.1f%%). This can cause SSSD watchdog terminations.", iowait))
				}
			}
		}
	})
}

func analyzeServices(dirPath string, report *ReportData) {
	var sssdStatusBlock []string
	inSssdStatus := false
	scanFiles(dirPath, []string{"systemd.txt"}, func(line string) {
		if strings.Contains(line, "sssd.service") && (strings.Contains(line, "Loaded:") || strings.HasPrefix(line, "●") || strings.Contains(line, "System Security Services Daemon")) {
			inSssdStatus = true
		}
		if inSssdStatus && strings.HasPrefix(line, "●") && !strings.Contains(line, "sssd.service") {
			inSssdStatus = false
		}
		if inSssdStatus {
			sssdStatusBlock = append(sssdStatusBlock, line)
		}

		if strings.Contains(line, "winbind.service") && strings.Contains(line, "Active: active") {
			report.WinbindService = constants.StatusRunning
			report.Problems = append(report.Problems, "Winbind is running. This can cause ID mapping conflicts with SSSD.")
		}
		if strings.Contains(line, "nscd.service") && strings.Contains(line, "Active: active") {
			report.NscdStatus = constants.StatusRunning
		}
	})

	for _, line := range sssdStatusBlock {
		if strings.Contains(line, "Active: active (running)") {
			report.SssdService = constants.StatusRunning
		}

		matches := reStatus.FindStringSubmatch(line)
		if len(matches) > 1 {
			switch matches[1] {
			case "1":
				report.Problems = append(report.Problems, "[SERVICE] sssd.service crashed with exit code 1 (Generic Error). Often a crash or a failed dependency.")
			case "2":
				report.Problems = append(report.Problems, "[SERVICE] sssd.service crashed with exit code 2 (Invalid Usage). Incorrect command-line arguments.")
			case "3":
				report.Problems = append(report.Problems, "[SERVICE] sssd.service exited with code 3 (Not Running). The service is dead, but a PID file might still exist.")
			case "4":
				report.Problems = append(report.Problems, "[SERVICE] sssd.service crashed with exit code 4 (Configuration Error). Check sssd.conf for typos or syntax errors.")
			case "6":
				report.Problems = append(report.Problems, "[SERVICE] sssd.service crashed with exit code 6 (Not Configured). Usually means /etc/sssd/sssd.conf is missing or has incorrect permissions (0600 for older versions, 0640 root:sssd for 2.10+).")
			case "7":
				report.Problems = append(report.Problems, "[SERVICE] sssd.service exited with code 7 (Not Running).")
			case "17":
				report.Problems = append(report.Problems, "[SERVICE] sssd.service crashed with exit code 17 (Bad File Permissions). Check ownership and mode of /etc/sssd/sssd.conf.")
			}
		}
	}

	if report.WinbindService == constants.StatusNotRunning {
		report.WinbindService = constants.StatusNotRunningDisabled
	}
	if report.NscdStatus == constants.StatusNotRunning {
		report.NscdStatus = constants.StatusDisabled
	}
}

func analyzeDiskSpace(dirPath string, report *ReportData) {
	scanFiles(dirPath, []string{"fs-diskio.txt", "storage.txt"}, func(line string) {
		fields := strings.Fields(line)
		if len(fields) >= 5 && strings.HasSuffix(fields[len(fields)-2], "%") {
			mount := fields[len(fields)-1]
			if mount == "/" || mount == "/var" {
				if usePct, err := strconv.Atoi(strings.TrimSuffix(fields[len(fields)-2], "%")); err == nil && usePct > 95 {
					report.Problems = append(report.Problems, fmt.Sprintf("[CRITICAL] Partition %s is %d%% full. SSSD requires disk space for cache and logs.", mount, usePct))
				}
			}
		}
	})
}

func analyzeMACDenials(dirPath string, report *ReportData) {
	var macDenials []string
	scanFiles(dirPath, []string{"security-apparmor.txt", "security-selinux.txt"}, func(line string) {
		lowerLine := strings.ToLower(line)
		if (strings.Contains(lowerLine, "denied") || strings.Contains(lowerLine, "denying")) && strings.Contains(lowerLine, "sssd") {
			if len(macDenials) < 5 {
				macDenials = append(macDenials, strings.TrimSpace(line))
			}
		}
	})

	if len(macDenials) > 0 {
		report.Problems = append(report.Problems, fmt.Sprintf("[WARNING] %s denials detected for SSSD. This can silently block authentication or cache access.", report.MACType))
		report.MACDenialExamples = deduplicateProblems(macDenials)
	}
}

func analyzeMACStatus(dirPath string, report *ReportData) {
	report.MACType = constants.StatusUnknownNone

	scanFiles(dirPath, []string{"boot.txt"}, func(line string) {
		if strings.Contains(line, "security=apparmor") {
			report.MACType = "AppArmor"
		} else if strings.Contains(line, "security=selinux") || strings.Contains(line, "selinux=1") {
			report.MACType = "SELinux"
		}
	})

	if report.MACType == constants.StatusUnknownNone {
		if anyFileContains(dirPath, []string{"security-apparmor.txt"}, "apparmor module is loaded") || anyFileContains(dirPath, []string{"security-apparmor.txt"}, "Active: active") {
			report.MACType = "AppArmor"
		} else if anyFileContains(dirPath, []string{"security-selinux.txt"}, "SELinux status:                 enabled") {
			report.MACType = "SELinux"
		}
	}
}
