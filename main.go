package main

import (
	"fmt"           // Formatting strings
	"io/ioutil"     // Reading/writing files and directories
	"log"           // Logging errors and fatal messages
	"os"            // OS functions (e.g., checking root privileges)
	"os/exec"       // Executing shell commands
	"path/filepath" // Constructing file paths
	"strings"       // String manipulation
)

// netplanDir: Directory where netplan YAML files are stored
// backupSuffix: Suffix added to make backup copies
// configFile: Name of the new netplan configuration file
const (
	netplanDir   = "/etc/netplan"
	backupSuffix = ".bak"
	configFile   = "01-static-network.yaml"
)

// checkRoot ensures the program is run with root privileges
func checkRoot() {
	if os.Geteuid() != 0 {
		log.Fatal("Error: This tool must be run as root.")
	}
}

// backupNetplan reads all .yaml/.yml files in netplanDir and copies
// them with a .bak suffix if a backup does not already exist
func backupNetplan() {
	files, err := ioutil.ReadDir(netplanDir)
	if err != nil {
		log.Fatalf("Failed to read netplan dir: %v", err)
	}
	for _, f := range files {
		ext := filepath.Ext(f.Name())
		if ext == ".yaml" || ext == ".yml" {
			src := filepath.Join(netplanDir, f.Name())
			dst := src + backupSuffix
			if _, err := os.Stat(dst); os.IsNotExist(err) {
				// Copy original file to backup
				if err := exec.Command("cp", src, dst).Run(); err != nil {
					log.Fatalf("Backup failed: %v", err)
				}
			}
		}
	}
}

// writeConfig generates a new netplan YAML file with the given parameters:
// iface: network interface (e.g., eth0)
// address: IP address (e.g., 192.168.3.51)
// cidr: subnet prefix length (e.g., 24)
// gateway: default gateway IP
// dns: list of DNS server IPs
func writeConfig(iface, address, cidr, gateway string, dns []string) {
	path := filepath.Join(netplanDir, configFile)
	// Quote each DNS entry for YAML syntax
	quoted := make([]string, len(dns))
	for i, d := range dns {
		quoted[i] = fmt.Sprintf("\"%s\"", d)
	}
	// Build YAML content
	content := fmt.Sprintf(`network:
  version: 2
  renderer: networkd
  ethernets:
    %s:
      dhcp4: no
      addresses: ["%s/%s"]
      gateway4: %s
      nameservers:
        addresses: [%s]
`, iface, address, cidr, gateway, strings.Join(quoted, ", "))
	// Write to file, overwriting if exists
	if err := ioutil.WriteFile(path, []byte(content), 0644); err != nil {
		log.Fatalf("Failed to write config: %v", err)
	}
}

// applyNetplan runs 'netplan apply' to activate the new network configuration
func applyNetplan() {
	if err := exec.Command("netplan", "apply").Run(); err != nil {
		log.Fatalf("netplan apply failed: %v", err)
	}
}

func main() {
	// Ensure correct number of arguments
	if len(os.Args) < 5 {
		fmt.Println("Usage: sudo fixip <interface> <ip-address> <cidr> <gateway> [<dns1[,dns2,...]>]")
		os.Exit(1)
	}
	iface := os.Args[1]
	address := os.Args[2]
	cidr := os.Args[3]
	gateway := os.Args[4]
	dns := []string{"8.8.8.8"} // Default DNS
	if len(os.Args) >= 6 {
		dns = strings.Split(os.Args[5], ",")
	}

	// 1. Check for root privileges
	checkRoot()
	// 2. Backup existing netplan configs
	backupNetplan()
	// 3. Write the static IP configuration
	writeConfig(iface, address, cidr, gateway, dns)
	// 4. Apply the new configuration
	applyNetplan()

	fmt.Printf("✅ Static IP set to %s/%s on %s\n", address, cidr, iface)
}
