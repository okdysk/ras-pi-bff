# ====== 設定 ======
# Raspberry Pi のアーキテクチャ
GOOS   ?= linux
GOARCH ?= arm64   # aarch64ならarm64, armv6l・armv7lならarm
GOARM  ?=         # arm64なら空, armなら6や7を指定
CGO_ENABLED ?= 0  # cgoを使わないなら0推奨（cgo=C言語のライブラリ）

# 出力バイナリ名
BINARY ?= bbs-server
# エントリーポイント
ENTRY  ?= server.go

# ====== ターゲット ======

# デフォルト: ビルド
build:
	@echo "Building for $(GOOS)/$(GOARCH)$(if $(GOARM),/$(GOARM),) ..."
	GOOS=$(GOOS) GOARCH=$(GOARCH) $(if $(GOARM),GOARM=$(GOARM),) CGO_ENABLED=$(CGO_ENABLED) \
		go build -o $(BINARY) $(ENTRY)
	@echo "Done: ./$(BINARY)"

# クリーン
clean:
	@rm -f $(BINARY)
	@echo "Cleaned"

# 現在のGoビルド設定を確認
env:
	go env GOOS GOARCH GOARM CGO_ENABLED