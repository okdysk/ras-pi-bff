# セットアップ手順

## [1] GOARCH / GOARM の調べ方

ラズパイにSSHして
```
yoki@raspberrypi:~ $ uname -m
aarch64
```
つまり、```GOARCH=arm64```だけを指定する。ラズパイ4とかだったらGOARMが出てくるからそれも含める。

## [2] makefileの編集、ローカルでクロスコンパイル

- makefileを編集
```
# Raspberry Pi のアーキテクチャ
GOOS   ?= linux
GOARCH ?= arm64   # aarch64ならarm64, armv6l・armv7lならarm
GOARM  ?=         # arm64なら空, armなら6や7を指定
CGO_ENABLED ?= 0  # cgoを使わないなら0推奨（cgo=C言語のライブラリ）

# 出力バイナリ名
BINARY ?= bbs-server
# エントリーポイント
ENTRY  ?= server.go
```

- コンパイル

Windowsなので、WSL（Ubuntu）経由でやってるよ。
```
$ make build
```

makefileで``` # 出力バイナリ名 BINARY ?= bbs-server ```と設定していれば、bbs-serverってバイナリファイルが生成されます。
```
$ file ./bbs-server
./bbs-server: ELF 64-bit LSB executable, ARM aarch64, version 1 (SYSV), statically linked, BuildID[sha1]=617e90fe394da34cec23c923de9b3100014424f6, with debug_info, not stripped
```

## [3] SCPでラズパイに転送

プロジェクトルートでラズパイ（yoki@192.168.40.185）に向かってSCP
```
> scp .\.env .\bbs-server yoki@192.168.40.185:/home/yoki/
```

ラズパイ側で
```
$ chmod +x ~/bbs-server
```

## [4] ラズパイで実行確認

```
yoki@raspberrypi:~ $ ./bbs-server 
2025/08/10 07:46:54 ✅ DB接続に成功しました
2025/08/10 07:46:54 connect to http://localhost:8080/ for GraphQL playground
```
http://192.168.40.185:8080/graphql にて、プレイグラウンドが出ればOK！

## [5] systemdでサービス化（ここからラズパイ設定の話）

### アプリファイルを移動
ユーザーホームにあると間違えて.envとか上書きしちゃうかもしれないし
```
$ sudo mkdir -p /opt/bbs-server
$ sudo chown yoki:yoki /opt/bbs-server
$ mv ~/bbs-server /opt/bbs-server/
$ mv ~/.env /opt/bbs-server/.env
```

### systemd ユニットファイル作成
```
$ sudo nano /etc/systemd/system/bbs-server.service
```
中身
```
[Unit]
Description=BBS GraphQL Server
After=network.target

[Service]
Type=simple
User=yoki
WorkingDirectory=/opt/bbs-server
ExecStart=/opt/bbs-server/bbs-server
Restart=always
RestartSec=3
EnvironmentFile=/opt/bbs-server/.env

[Install]
WantedBy=multi-user.target
```

### 読み込み & 有効化 & 起動
```
$ sudo systemctl daemon-reload
$ sudo systemctl enable bbs-server
$ sudo systemctl start bbs-server
```
### 動作確認
```
$ systemctl status bbs-server
```

### 停止や再起動
```
$ sudo systemctl stop bbs-server
$ sudo systemctl restart bbs-server
```
