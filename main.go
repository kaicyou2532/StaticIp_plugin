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

// ======= デフォルト設定（簡易モード用） =======
const (
    defaultGateway  = "192.168.3.1"  // ゲートウェイ
    defaultCIDR     = "24"           // ネットマスク255.255.255.0相当
    defaultDNS      = "8.8.8.8"      // DNSサーバ
    netplanDir      = "/etc/netplan" // netplan ディレクトリ
    configFilename  = "01-static-network.yaml"
    backupSuffix    = ".bak"
)

// usage は使い方を表示して終了します
func usage() {
    fmt.Println("Usage:")
    fmt.Println("  フルモード: sudo fixip <interface> <ip-address> <cidr> <gateway> [<dns1[,dns2,...]>]")
    fmt.Println("  簡易モード: sudo fixip <interface> <ip-address>")
    os.Exit(1)
}

func main() {
    // ── 1. root 権限チェック
    if os.Geteuid() != 0 {
        fmt.Fprintln(os.Stderr, "root 権限で実行してください")
        os.Exit(1)
    }

    // ── 2. 引数解析
    args := os.Args[1:]
    var iface, ip, cidr, gateway, dnsList string

    switch len(args) {
    case 2:
        // 簡易モード：<IFACE> <IP> のみ
        iface = args[0]
        ip = args[1]
        cidr = defaultCIDR
        gateway = defaultGateway
        dnsList = defaultDNS

    case 4, 5:
        // フルモード：<IFACE> <IP> <CIDR> <GW> [DNS]
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

    // ── 3. 既存 netplan 設定ファイルをバックアップ
    backupNetplanFiles()

    // ── 4. 新しい YAML コンテンツを生成
    yamlContent := generateNetplanYAML(iface, ip, cidr, gateway, dnsList)

    // ── 5. 設定ファイルを書き込み（0600 パーミッション）
    configPath := filepath.Join(netplanDir, configFilename)
    if err := ioutil.WriteFile(configPath, []byte(yamlContent), 0600); err != nil {
        fmt.Fprintf(os.Stderr, "設定書き込みエラー: %v\n", err)
        os.Exit(1)
    }
    // 念のため再度パーミッションを設定
    if err := os.Chmod(configPath, 0600); err != nil {
        fmt.Fprintf(os.Stderr, "パーミッション設定エラー: %v\n", err)
        os.Exit(1)
    }

    // ── 6. netplan apply 実行
    if err := runCommand("netplan", "apply"); err != nil {
        fmt.Fprintf(os.Stderr, "netplan apply エラー: %v\n", err)
        os.Exit(1)
    }

    // ── 7. 完了メッセージ
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

// generateNetplanYAML は netplan 用 YAML コンテンツを生成します
func generateNetplanYAML(iface, ip, cidr, gateway, dnsList string) string {
    dnsEntries := `["` + strings.ReplaceAll(dnsList, ",", `","`) + `"]`
    return fmt.Sprintf(`network:
  version: 2
  renderer: networkd
  ethernets:
    %s:
      dhcp4: no
      addresses:
        - "%s/%s"
      routes:
        - to: 0.0.0.0/0
          via: %s
      nameservers:
        addresses: %s
`, iface, ip, cidr, gateway, dnsEntries)
}

// copyFile は src を dst にコピーします
func copyFile(src, dst string) error {
    data, err := ioutil.ReadFile(src)
    if err != nil {
        return err
    }
    return ioutil.WriteFile(dst, data, 0600)
}

// runCommand は外部コマンドを同期実行します
func runCommand(name string, args ...string) error {
    cmd := exec.Command(name, args...)
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    return cmd.Run()
}
