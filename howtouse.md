 StaticIp Plugin 設定手順

このドキュメントでは、StaticIp_plugin を使って Ubuntu サーバーでの固定 IP 設定をワンコマンドで行う手順をまとめています。

## 1. 必要パッケージのインストール
```bash
sudo apt update
sudo apt install -y golang-go
```
（または最新の Go が必要な場合）
```bash
sudo snap install go --classic
```

## 2. リポジトリのクローン
```bash
git clone https://github.com/youruser/StaticIp_plugin.git
cd StaticIp_plugin
```

## 3. Go モジュールの初期化
```bash
go mod init StaticIp_plugin
go mod tidy
```

## 4. ビルド
```bash
go build -o fixip
```

## 5. インストール
```bash
sudo mv fixip /usr/local/bin/
```

## 6. 使い方
### フル設定モード
```bash
sudo fixip <インターフェース名> <IP アドレス> <プレフィックス長> <ゲートウェイ> <DNS リスト>
# 例:
sudo fixip eth0 192.168.3.51 24 192.168.3.1 8.8.8.8,1.1.1.1
```

### 簡易設定モード
```bash
sudo fixip <インターフェース名> <IP アドレス>
# ゲートウェイ: 192.168.3.1
# ネットマスク: 255.255.255.0 （プレフィックス 24）
# DNS: 8.8.8.8
```

以上で、StaticIp_plugin による固定 IP 設定がワンコマンドで完了します。
