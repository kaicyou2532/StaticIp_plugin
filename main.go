// main.go
package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// デフォルト設定（簡易モード時に使用）
const (
	defaultGateway = "192.168.3.1"
	defaultCIDR    = "24" // ネットマスク255.255.255.0相当
	defaultDNS     = "8.8.8.8"
	netplanDir     = "/etc/netplan"
	configFilename = "01-static-network.yaml"
	backupSuffix   = ".bak"
)

func usage() {
	fmt.Println("Usage:")
	fmt.Println("  フルモード: sudo fixip <interface> <ip-address> <cidr> <gateway> [<dns1[,dns2,...]>]")
	fmt.Println("  簡易モード: sudo fixip <interface> <ip-address>")
	os.Exit(1)
}

func main() {
	// root 権限チェック
	if os.Geteuid() != 0 {
		fmt.Fprintln(os.Stderr, "root 権限で実行してください")
		os.Exit(1)
	}

	args := os.Args[1:]
	var iface, ip, cidr, gateway, dnsList string

	switch len(args) {
	case 2:
		// 簡易モード
		iface = args[0]
		ip = args[1]
		cidr = defaultCIDR
		gateway = defaultGateway
		dnsList = defaultDNS

	case 4, 5:
		// フルモード
		iface = args[0]
		ip = args[1]
		cidr = args[2]
		gateway = args[3]
		if len(args) == 5 && args[4] != "" {
			dnsList = args[4]
		} else {
			dnsList = defaultDNS
		}

	default:
		usage()
	}

	// netplan ディレクトリ内の既存 YAML をバックアップ
	backupNetplanFiles()

	// 新設定ファイルの生成
	yaml := generateNetplanYAML(iface, ip, cidr, gateway, dnsList)

	// ファイル書き込み
	path := filepath.Join(netplanDir, configFilename)
	if err := ioutil.WriteFile(path, []byte(yaml), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "設定書き込みエラー: %v\n", err)
		os.Exit(1)
	}

	// netplan apply 実行
	if err := runCommand("netplan", "apply"); err != nil {
		fmt.Fprintf(os.Stderr, "netplan apply エラー: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ %s に %s/%s を設定しました (GW: %s, DNS: %s)\n",
		iface, ip, cidr, gateway, dnsList)
}

// backupNetplanFiles は /etc/netplan/*.yaml を .bak としてバックアップします
func backupNetplanFiles() {
	files, err := filepath.Glob(filepath.Join(netplanDir, "*.yaml"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "バックアップファイル探索エラー: %v\n", err)
		os.Exit(1)
	}
	for _, f := range files {
		bak := f + backupSuffix
		if _, err := os.Stat(bak); os.IsNotExist(err) {
			if err := copyFile(f, bak); err != nil {
				fmt.Fprintf(os.Stderr, "バックアップエラー (%s): %v\n", f, err)
				os.Exit(1)
			}
		}
	}
}

// generateNetplanYAML は netplan 用の YAML コンテンツを組み立てます
func generateNetplanYAML(iface, ip, cidr, gateway, dnsList string) string {
	// DNS リストを ["a","b",...] の形式に変換
	dnsEntries := `["` + strings.ReplaceAll(dnsList, ",", `","`) + `"]`

	return fmt.Sprintf(`network:
  version: 2
  renderer: networkd
  ethernets:
    %s:
      dhcp4: no
      addresses: ["%s/%s"]
      gateway4: %s
      nameservers:
        addresses: %s
`, iface, ip, cidr, gateway, dnsEntries)
}

// copyFile は src を dst にバイトコピーします
func copyFile(src, dst string) error {
	data, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(dst, data, 0644)
}

// runCommand は外部コマンドを同期実行します
func runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
