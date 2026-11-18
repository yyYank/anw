# anw

**anyway**、なんとかする。

日常の細かい面倒をとりあえず片づけるための、Go製のコマンド集です。
散らかったファイルを整頓し、埋もれた設定を見つけ、邪魔なプロセスを止めます。
ノイズを減らして、探すべきものを探しやすくするための道具です。

## インストール

```sh
go install github.com/yyYank/anw@latest
```

## コマンド

| コマンド | 内容 |
| --- | --- |
| `anw tidy` | ファイルを更新日時ごとのディレクトリに振り分ける |
| `anw env` | 配下の env ファイルを探して、閲覧・編集する |
| `anw port` | ポートを使っているプロセスを確認して止める |
| `anw undo` | 直前のファイル操作を元に戻す |

---

### anw tidy

カレントディレクトリ直下のファイルを、更新日時の `yyyy-mm-dd/` ディレクトリに振り分けます。

```sh
anw tidy
anw tidy --by date
```

```
.
├── memo.md            (2026-10-27 更新)
├── screenshot.png     (2026-10-29 更新)
└── report.pdf         (2026-10-29 更新)
```

↓

```
.
├── 2026-10-27/
│   └── memo.md
└── 2026-10-29/
    ├── report.pdf
    └── screenshot.png
```

- 対象はカレントディレクトリ直下のファイルのみです。ディレクトリは移動しません
- 隠しファイルは対象外です
- 移動先に同名のファイルがある場合は、`report (1).pdf` のように連番を付けます
- 実行結果は `anw undo` で元に戻せます

### anw env

カレントディレクトリ配下の env ファイルを検出し、TUI で一覧・閲覧・編集します。

```sh
anw env --view   # 一覧と内容のプレビュー
anw env --edit   # 選んだファイルを編集
```

- `.env` および `.env.*`（`.env.local`、`.env.production` など）を検出します
- `.git`、`node_modules`、`vendor` 配下は探索しません
- `--edit` は `$EDITOR` でファイルを開きます。見つかったファイルが1つだけなら、選択を挟まずに開きます

### anw port

LISTEN 中のポートと、それを使っているプロセスを表示します。

```sh
anw port         # 一覧から選んで kill
anw port 3000    # 3000番を使っているプロセスを kill
```

- 一覧にはポート番号、PID、プロセス名、コマンドを表示します
- kill はまず SIGTERM を送ります

### anw undo

anw が直前に行ったファイル操作を、まとめて元に戻します。

```sh
anw undo
```

- `anw tidy` による移動が対象です
- 操作の履歴は、既定では `~/.anw/` に保存されます（[設定](#設定)で変更できます）

## 設定

`~/.config/anw/config.toml` で設定できます。

```toml
# 操作の履歴を保存するディレクトリ
history_dir = "~/.anw"
```

- 値は toml → 環境変数 `ANW_HISTORY_DIR` → `~/.anw` の順で決まります
- 先頭の `~` と、`$HOME` などの環境変数は展開されます
- 読み込む設定ファイルは `--config` で指定できます

```sh
anw --config ./config.toml tidy
```

## License

Apache License 2.0
