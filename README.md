## ras-pi-bff
ラズパイで動いているBFF。

このGraphQLサーバーは Go + gqlgen を使って構築されています。

運用の際のビルド環境（WSL）：
```
yoki@DESKTOP-V6KQ6K9:/mnt/g/data/Git/Go/ras-pi-bff$ go version
go version go1.24.5 linux/amd64
```

### API構成図
```
[Frontend] --> [GraphQL BFF (Go)：このリポジトリ] --> [DB (MariaDB)]
```

## ドキュメント
（ペライチ）

https://docs.google.com/spreadsheets/d/12LV1oEodxzFdJBy3-y0C5DSvx9Mzfyd-Ag5P17ENl7Q/edit?gid=0#gid=0

他、docディレクトリに色々入れています。

### セットアップ手順
[docs/setup.md](docs/setup.md) 


### リクエスト＆レスポンスの例
[docs/playground.txt](docs/playground.txt) 

### ファイル構成
[structure.txt](structure.txt) 
