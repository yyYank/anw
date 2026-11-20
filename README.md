# anw

**anyway**、なんとかする。

日常の細かい面倒をとりあえず片づけるための、Go製のコマンド集です。
散らかったファイルを整頓し、埋もれた設定を見つけ、邪魔なプロセスを止めます。
ノイズを減らして、探すべきものを探しやすくするための道具です。

## インストール

```sh
go install github.com/yyYank/anw@latest
```

## 使い方

サブコマンドは動詞、対象はフラグで指定します。

```sh
anw <動詞> --<対象>
```

| 動詞 | 内容 | 対象 |
| --- | --- | --- |
| `view` | 見る | `--env` `--port` `--branch` |
| `edit` | 編集する | `--env` |
| `kill` | プロセスを止める | `--port` |
| `delete` | 削除する | `--branch` |
| `tidy` | ファイルを振り分ける | |
| `undo` | 直前の操作を元に戻す | |

`view` は見るだけで、何も変更しません。

### 対象の省略

対象を省略して動詞だけ打つと、その動詞で使える対象から選べます。選んだあとは、フラグを付けた場合と同じ動作になります。

```sh
anw view            # env / port / branch から選ぶ
anw view --env      # 選択を挟まず直行
```

```
view: 何を見ますか？

> env      env ファイルの一覧と内容のプレビュー
  port     LISTEN 中のポートの一覧
  branch   削除できるブランチの一覧
```

| キー | 動き |
| --- | --- |
| `↑` `↓` / `j` `k` | 移動 |
| `Enter` | 決定 |
| `q` / `Esc` | 終了 |

- 対象が1つしかない動詞（`edit`、`kill`、`delete`）は、選択を挟まずにその対象へ進みます
- `anw` だけを打つと、動詞と対象の対応表を出します
- その動詞で使えない対象を指定すると、使える対象を表示します（例: `anw kill --env` → `kill で使えるのは --port です`）

---

### anw view

対象を一覧・表示します。

```sh
anw view --env          # env ファイルの一覧と内容のプレビュー
anw view --port         # LISTEN 中のポートの一覧
anw view --port 3000    # 3000番を使っているプロセス
anw view --branch       # 削除できるブランチの一覧
```

**`--env`**

- カレントディレクトリ配下の `.env` および `.env.*`（`.env.local`、`.env.production` など）を検出します
- `.git`、`node_modules`、`vendor` 配下は探索しません
- 一覧は検索で絞り込めます（`anw edit --env` も同じ画面です）

| キー | 動き |
| --- | --- |
| `/` | パスで絞り込む。文字が順に含まれていれば一致します（`front`、`apps/front`、`afl` で `apps/front/.env.local` に一致） |
| `Ctrl+G` | 中身で絞り込む。大文字小文字を区別せず、含んでいれば一致します（`API_PORT=8080` は `api`、`PORT`、`8080` で一致） |
| `Enter`（入力中） | 絞り込みを残して入力を終える |
| `Esc`（入力中） | 絞り込みを解除して入力を終える |
| `↑` `↓` / `j` `k` | 移動（入力中は `↑` `↓` / `Ctrl+P` `Ctrl+N`） |
| `q` / `Esc` | 終了 |

**`--port`**

- ポート番号、PID、プロセス名、コマンドを表示します

**`--branch`**

- デフォルトブランチにマージ済みのローカルブランチと、リモートで削除済みのローカルブランチを表示します
- 現在のブランチは含みません
- デフォルトブランチは `origin/HEAD` から決めます。fetch はしないので、必要なら先に `git fetch --prune` してください

### anw edit

対象を選んで編集します。

```sh
anw edit --env
```

- 対象は `anw view --env` と同じです
- `$EDITOR` でファイルを開きます
- 見つかったファイルが1つだけなら、選択を挟まずに開きます

### anw kill

プロセスを止めます。

```sh
anw kill --port         # 一覧から選んで kill
anw kill --port 3000    # 3000番を使っているプロセスを kill
```

- SIGTERM を送ります

### anw delete

対象を選んで削除します。

```sh
anw delete --branch           # 一覧から選んで削除
anw delete --force --branch   # マージされていなくても強制削除
```

- 対象は `anw view --branch` と同じです
- Space で複数選べます。何も選ばずに Enter を押すと、カーソル位置のブランチを削除します
- 通常は `git branch --delete` で削除するため、マージされていないブランチ（squash マージしたものなど）は削除されません
- `--force` を付けると `git branch --delete --force` で強制削除します

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
