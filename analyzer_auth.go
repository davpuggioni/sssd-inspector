// analyzer_auth.go
package main

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

func analyzeDNS(dirPath string, report *ReportData) {
	dnsStatusMap := make(map[string]string)

	scanFiles(dirPath, []string{"network.txt"}, func(line string) {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# Connectivity Test, DNS Server") {
			parts := strings.Split(line, ":")
			if len(parts) == 2 {
				ipParts := strings.Fields(parts[0])
				if len(ipParts) > 0 {
					dnsStatusMap[ipParts[len(ipParts)-1]] = strings.TrimSpace(parts[1])
				}
			}
		}
	})

	resolvContent := extractSection(dirPath, "network.txt", "# /etc/resolv.conf")
	if resolvContent == "" {
		resolvContent = extractSection(dirPath, "etc.txt", "# /etc/resolv.conf")
	}

	if resolvContent != "" {
		for _, line := range strings.Split(resolvContent, "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "nameserver") {
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					ip := parts[1]
					statusStr := "Unknown/Not Tested"
					if status, ok := dnsStatusMap[ip]; ok {
						statusStr = status
					}
					if statusStr != "Success" && statusStr != "Unknown/Not Tested" {
						report.Problems = append(report.Problems, fmt.Sprintf("[WARNING] DNS Server %s ping test failed (Status: %s). Note: ICMP may be blocked in cloud environments like Azure.", ip, statusStr))
					}
					report.Nameservers = append(report.Nameservers, fmt.Sprintf("%s (%s)", ip, statusStr))
				}
			}
			if strings.HasPrefix(line, "search") {
				report.SearchDomain = strings.TrimPrefix(line, "search ")
			}
		}
	}
	if len(report.Nameservers) == 0 {
		report.Problems = append(report.Problems, "No nameservers found in /etc/resolv.conf. DNS resolution will fail.")
	}
}

func analyzeTime(dirPath string, report *ReportData) {
	timeFiles := []string{"systemd.txt", "ntp.txt"}

	hasChronyd := anyFileContains(dirPath, timeFiles, "chronyd.service")
	hasNtpd := anyFileContains(dirPath, timeFiles, "ntpd.service")
	isActive := anyFileContains(dirPath, timeFiles, "Active: active (running)")

	if hasChronyd && isActive {
		report.TimeService = "chronyd (Running)"
	} else if hasNtpd && isActive {
		report.TimeService = "ntpd (Running)"
	}

	isSynced := anyFileContains(dirPath, timeFiles, "System clock synchronized: yes") ||
		(anyFileContains(dirPath, timeFiles, "^*") && anyFileContains(dirPath, timeFiles, "377"))

	if anyFileContains(dirPath, timeFiles, "chronyc sources") {
		if !anyFileContains(dirPath, timeFiles, " 377 ") {
			report.Problems = append(report.Problems, "[NTP] Chrony reachability is not 377. Time servers may be unreachable, risking Kerberos authentication failure.")
		}
		if anyFileContains(dirPath, timeFiles, "#* PHC0") {
			report.Problems = append(report.Problems, "[NTP] System is synchronized only to a local clock (PHC0) instead of a network time server.")
		}
	}

	scanFiles(dirPath, timeFiles, func(line string) {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "System time") && strings.Contains(line, "seconds") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				fields := strings.Fields(parts[1])
				if len(fields) > 0 {
					if offset, err := strconv.ParseFloat(fields[0], 64); err == nil {
						if math.Abs(offset) > 300.0 {
							report.Problems = append(report.Problems, fmt.Sprintf("[CRITICAL] System time offset is %.2f seconds. Kerberos requires a maximum skew of 300 seconds (5 minutes). Active Directory logins will fail.", math.Abs(offset)))
						}
					}
				}
			}
		}
	})

	if report.TimeService != "Not Running / Unknown" {
		if isSynced {
			report.TimeService += " & Synchronized"
		} else {
			report.TimeService += " (Not Synchronized)"
			report.Problems = append(report.Problems, "Time service is running, but clock does not appear synchronized. Kerberos requires synchronized time.")
		}
	} else {
		report.Problems = append(report.Problems, "No active Time Synchronization service found. Kerberos/AD requires synchronized time.")
	}
}

func analyzePAM(dirPath string, report *ReportData) {
	pam := readFileSafe(dirPath, "pam.txt")
	if strings.Contains(pam, "FORCE_OPTION_PAM=1") || strings.Contains(pam, "General Data Protection Regulation") {
		report.PamGDPRRestricted = true
		report.Problems = append(report.Problems, "[WARNING] PAM data is restricted (GDPR). Please collect a new supportconfig using: FORCE_OPTION_PAM=1 supportconfig")
	} else if strings.Contains(pam, "pam_sss.so") {
		report.PamSssInstalled = true
	}

	if !report.PamSssInstalled && !report.PamGDPRRestricted {
		report.Problems = append(report.Problems, "PAM configuration does not include pam_sss.so. SSSD authentication will not be triggered.")
	}
}

func analyzeNSSwitch(dirPath string, report *ReportData) {
	nssContent := readFileSafe(dirPath, "nsswitch.conf")
	if nssContent == "" {
		nssContent = extractSection(dirPath, "etc.txt", "# /etc/nsswitch.conf")
	}
	if nssContent == "" {
		nssContent = extractSection(dirPath, "sssd.txt", "# /etc/nsswitch.conf")
	}

	if nssContent != "" {
		hasPasswd, hasGroup := false, false
		for _, line := range strings.Split(nssContent, "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "#") {
				continue
			}
			isPasswd, isGroup := strings.HasPrefix(line, "passwd:"), strings.HasPrefix(line, "group:")

			if isPasswd || isGroup {
				if strings.Contains(line, "sss") {
					if isPasswd {
						hasPasswd = true
					}
					if isGroup {
						hasGroup = true
					}
					parts := strings.Fields(line)
					sssIdx, filesIdx := -1, -1
					for i, p := range parts {
						if p == "sss" {
							sssIdx = i
						}
						if p == "files" || p == "compat" {
							filesIdx = i
						}
					}
					if sssIdx != -1 && filesIdx != -1 && sssIdx < filesIdx {
						report.Problems = append(report.Problems, fmt.Sprintf("[WARNING] In /etc/nsswitch.conf, 'sss' is listed before 'files'/'compat' for '%s'. This can lock out local root/system accounts if AD is unreachable.", strings.TrimSuffix(parts[0], ":")))
					}
				}
			}
		}
		report.NsswitchValid = (hasPasswd && hasGroup)
		if !report.NsswitchValid {
			report.Problems = append(report.Problems, "/etc/nsswitch.conf is missing 'sss' in passwd or group modules.")
		}
	}
}

func analyzeHosts(dirPath string, report *ReportData) {
	hasLocalhost := false
	scanFiles(dirPath, []string{"hosts"}, func(line string) {
		if strings.Contains(line, "127.0.0.1") {
			hasLocalhost = true
		}
	})
	if !hasLocalhost {
		report.HostsIssues = append(report.HostsIssues, "Missing 127.0.0.1 loopback entry")
		report.Problems = append(report.Problems, "Malformed /etc/hosts file.")
	}
}

func analyzeNSCD(dirPath string, report *ReportData) {
	scanFiles(dirPath, []string{"nscd.conf"}, func(line string) {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "enable-cache") {
			parts := strings.Fields(line)
			if len(parts) >= 3 && parts[2] == "yes" && (parts[1] == "passwd" || parts[1] == "group" || parts[1] == "netgroup") {
				report.NscdCaching = append(report.NscdCaching, parts[1])
				report.Problems = append(report.Problems, fmt.Sprintf("nscd is caching '%s' which can conflict with SSSD.", parts[1]))
			}
		}
	})
}

func analyzePackages(dirPath string, report *ReportData) {
	inPackageBlock := false
	scanFiles(dirPath, []string{"rpm.txt"}, func(line string) {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#==[") && inPackageBlock {
			inPackageBlock = false
		}
		if strings.HasPrefix(line, "sssd ") || strings.HasPrefix(line, "sssd-") {
			inPackageBlock, report.SssdInstalled = true, true
			report.SSSDPackages = append(report.SSSDPackages, line)
		}
	})
	report.SSSDPackages = deduplicateProblems(report.SSSDPackages)
}

func analyzeSSSDVersionAge(report *ReportData) {
	major, minor := getSSSDVersion(report.SSSDPackages)
	if major == 1 {
		report.Problems = append(report.Problems, fmt.Sprintf("[DEPRECATION] Installed SSSD version is %d.%d. The 1.x series is extremely outdated (last upstream release in 2020) and End-of-Life. Consider upgrading your OS or packages.", major, minor))
	} else if major == 2 && minor < 8 {
		report.Warnings = append(report.Warnings, fmt.Sprintf("[MAINTENANCE] Installed SSSD version is 2.%d. Upstream SSSD is currently actively releasing 2.10+ and 2.12+. If you are experiencing unexpected bugs, check for available OS package updates.", minor))
	}
}

func analyzeSSSDFilePermissions(dirPath string, report *ReportData) {
	major, minor := getSSSDVersion(report.SSSDPackages)
	if major == 0 {
		return
	}

	expectedConfOwner, expectedConfGroup, expectedConfPerms := "root", "root", "-rw-------" // 0600
	expectedVarOwner, expectedVarGroup := "root", "root"

	isSles15SP7 := strings.Contains(report.SLESRlease, "15 SP7") || strings.Contains(report.SLESRlease, "15-SP7")

	if major > 2 || (major == 2 && minor >= 10) {
		if isSles15SP7 {
			expectedConfOwner, expectedConfGroup, expectedConfPerms = "root", "root", "-rw-------"
			expectedVarOwner, expectedVarGroup = "root", "root"
		} else {
			expectedConfOwner, expectedConfGroup, expectedConfPerms = "root", "sssd", "-rw-r-----"
			expectedVarOwner, expectedVarGroup = "sssd", "sssd"
		}
	}

	foundConfErr := false
	varDirMistakes := 0
	inVarLibSss := false
	foundSssdUserInVar := false

	scanFiles(dirPath, []string{"sssd.txt", "etc.txt"}, func(line string) {
		lineTrimmed := strings.TrimSpace(line)
		if len(lineTrimmed) == 0 {
			return
		}

		if !foundConfErr && strings.HasPrefix(lineTrimmed, "-") && strings.HasSuffix(lineTrimmed, "sssd.conf") {
			fields := strings.Fields(lineTrimmed)
			if len(fields) >= 8 {
				perms, owner, group := fields[0], fields[2], fields[3]
				if owner != expectedConfOwner || group != expectedConfGroup {
					report.Problems = append(report.Problems, fmt.Sprintf("CONFIGURATION ERROR: sssd.conf is owned by '%s:%s', but installed SSSD version %d.%d strictly requires '%s:%s'. The service will fail to start.", owner, group, major, minor, expectedConfOwner, expectedConfGroup))
					foundConfErr = true
				}
				if !foundConfErr && perms != expectedConfPerms {
					report.Problems = append(report.Problems, fmt.Sprintf("CONFIGURATION ERROR: sssd.conf has incorrect permissions '%s'. SSSD version %d.%d requires exactly '%s' or it will refuse to start.", perms, major, minor, expectedConfPerms))
					foundConfErr = true
				}
			}
		}

		if strings.HasPrefix(lineTrimmed, "#==[") {
			inVarLibSss = false
		} else if strings.HasPrefix(lineTrimmed, "/") && strings.HasSuffix(lineTrimmed, ":") {
			inVarLibSss = strings.Contains(lineTrimmed, "/var/lib/sss")
		}

		if inVarLibSss {
			c := lineTrimmed[0]
			if c == '-' || c == 'd' || c == 's' || c == 'p' || c == 'l' || c == 'c' || c == 'b' {
				fields := strings.Fields(lineTrimmed)
				if len(fields) >= 8 {
					owner, group := fields[2], fields[3]
					if owner != expectedVarOwner || group != expectedVarGroup {
						varDirMistakes++
						if owner == "sssd" || group == "sssd" {
							foundSssdUserInVar = true
						}
					}
				}
			}
		}
	})

	if varDirMistakes > 0 {
		if isSles15SP7 && foundSssdUserInVar {
			report.Problems = append(report.Problems, fmt.Sprintf("CONFIGURATION ERROR: %d files or directories in /var/lib/sss/ are incorrectly owned. On SLES 15 SP7, if the /var/lib/sss directory permissions are still assigned to the sssd user, it is recommended to upgrade to a version of sssd later than 2.10.2-150700.9.17.1. Earlier versions may exhibit regressions when running in unprivileged mode. Ensure that ownership is reverted to root:root.", varDirMistakes))
		} else {
			report.Problems = append(report.Problems, fmt.Sprintf("CONFIGURATION ERROR: %d files or directories in /var/lib/sss/ are incorrectly owned. SSSD version %d.%d strictly requires them to be owned by '%s:%s'. Please run 'chown -R %s:%s /var/lib/sss/' to fix.", varDirMistakes, major, minor, expectedVarOwner, expectedVarGroup, expectedVarOwner, expectedVarGroup))
		}
	}
}
