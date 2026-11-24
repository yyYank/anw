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
| `view` | 見る | `--env` `--port` `--branch` `--worktree` `--stash` `--ai` `--dotfiles` `--snippet` |
| `edit` | 編集する | `--env` `--dotfiles` `--snippet` |
| `kill` | プロセスを止める | `--port` |
| `delete` | 削除する | `--branch` `--worktree` `--stash` |
| `tidy` | ファイルを振り分ける | |
| `undo` | 直前の操作を元に戻す | |
| `run` | snippet を実行する | `--snippet` |

`view` は見るだけで、何も変更しません。

### 対象の省略

対象を省略して動詞だけ打つと、その動詞で使える対象から選べます。選んだあとは、フラグを付けた場合と同じ動作になります。

```sh
anw view            # env / port / branch / worktree / stash / ai / dotfiles / snippet から選ぶ
anw view --env      # 選択を挟まず直行
```

```
view: 何を見ますか？

> env      env ファイルの一覧と内容のプレビュー
  port     LISTEN 中のポートの一覧
  branch   削除できるブランチの一覧
  worktree 消してよい worktree の一覧
  stash    stash の一覧
  ai       コーディングエージェントのセッション
  dotfiles ホームの設定ファイルの一覧と内容のプレビュー
  snippet  snippet の一覧
```

| キー | 動き |
| --- | --- |
| `↑` `↓` / `j` `k` | 移動 |
| `Enter` | 決定 |
| `q` / `Esc` | 終了 |

- 対象が1つしかない動詞（`kill`、`run`）は、選択を挟まずにその対象へ進みます
- `anw` だけを打つと、動詞と対象の対応表を出します
- その動詞で使えない対象を指定すると、使える対象を表示します（例: `anw kill --env` → `kill で使えるのは --port です`）

一覧の画面で選ばずに `q` / `Esc` でやめると、終了コード 130 で終わります。

---

### anw view

対象を一覧・表示します。

```sh
anw view --env          # env ファイルの一覧と内容のプレビュー
anw view --port         # LISTEN 中のポートの一覧
anw view --port 3000    # 3000番を使っているプロセス
anw view --branch       # 削除できるブランチの一覧
anw view --worktree     # 消してよい worktree の一覧
anw view --stash        # stash の一覧
anw view --ai           # コーディングエージェントのセッションを全プロジェクト横断で一覧
anw view --ai .         # 今いるリポジトリ（別の worktree を含む）のセッションだけ
anw view --dotfiles     # ホームの設定ファイルの一覧と内容のプレビュー
anw view --snippet      # snippet の一覧と中身
anw view --snippet hoge # snippet hoge の中身
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

**`--worktree`**

- ディレクトリがもう消えた worktree と、ブランチがマージ済み・リモート削除済みの worktree を、理由付きで表示します
- メインの worktree と、今いる worktree は含みません

**`--stash`**

- stash をすべて、元のブランチと何日前かを付けて表示します
- 元のブランチがもうないものには `(ブランチなし)` と付けます

**`--ai`**

- Claude Code（`~/.claude/projects/`）と Codex（`~/.codex/sessions/`）のセッションを、最終更新の新しい順に一覧します
- 一覧には最終更新日時、エージェントの種類、作業ディレクトリ、ブランチ、最初の発言を表示します
- 作業ディレクトリがもう存在しないセッションには `(削除済み)` と付けます
- プレビューには発言だけを表示し、ツールの呼び出しは `[tool] 名前` の一行に畳みます
- `/` で作業ディレクトリのパス、`Ctrl+G` で会話の中身を絞り込めます
- `Enter` を押すと、そのセッションを再開するコマンド（`cd '<作業ディレクトリ>' && claude --resume <id>` など）を出力します。実行はしません
- `.` を付けると、git の共通ディレクトリで判定して、カレントと同じリポジトリのセッションだけにします

**`--dotfiles`**

- ホーム直下のドットファイル（`.zshrc`、`.gitconfig` など）と、`~/.config` 配下の3階層までを検出します
- ホーム直下のドットディレクトリ（`~/.ssh` など）の中は探索しません
- シンボリックリンクはリンク先をたどります
- テキストでないファイル、1MB を超えるファイル、`cache`・`Cache`・`logs` ディレクトリの中は対象外です
- 画面と操作は `--env` と同じです（`/` でパス、`Ctrl+G` で中身を絞り込み）
- 管理はしません。探して見る・開くだけです

**`--snippet`**

- [設定](#設定) に書いた snippet の、ステップごとの dir と実行するコマンドを表示します

### anw edit

対象を選んで編集します。

```sh
anw edit --env
anw edit --dotfiles
anw edit --snippet morning   # snippet morning を書く。なければ新しく作る
anw edit --snippet           # 保存済みの snippet から選んで書く
```

- 対象は `anw view --env`、`anw view --dotfiles` と同じです
- `$EDITOR` でファイルを開きます
- 見つかったファイルが1つだけなら、選択を挟まずに開きます

**`--snippet`**

- snippet を書く画面を開きます。書き方は [snippet](#snippet) を見てください
- 保存すると、設定ファイルの隣の `snippets/名前`（既定では `~/.config/anw/snippets/名前`）に書きます。もう一度開くと、保存した内容から書き足せます
- 保存のたびに `anw run` と同じ確認をし、実行できない内容は保存せずにエラーを表示します
- `config.toml` に書いた snippet は編集できません
- 入力モードで始まり、vim のように操作します

| キー | 動き |
| --- | --- |
| `Ctrl+S` | 保存（どのモードでも） |
| `Esc` | ノーマルモードへ |
| `i` / `a` / `o` | 入力モードへ（カーソルの前 / 後 / 下に行を足して） |
| `h` `j` `k` `l` / 矢印 | 移動 |
| `0` / `$` | 行頭 / 行末へ |
| `x` | 1文字消す |
| `:w` / `:wq` / `:q` / `:q!` | 保存 / 保存して終了 / 終了（未保存の変更があれば止める） / 破棄して終了 |
| `Ctrl+C` | 保存せずに終了 |

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
anw delete --worktree         # 一覧から選んで削除
anw delete --force --worktree # 変更が残っていても強制削除
anw delete --stash            # 一覧から選んで削除
```

- 対象は `anw view` の同じ対象と同じです
- Space で複数選べます。何も選ばずに Enter を押すと、カーソル位置のものを削除します

**`--branch`**

- 通常は `git branch --delete` で削除するため、マージされていないブランチ（squash マージしたものなど）は削除されません
- `--force` を付けると `git branch --delete --force` で強制削除します
- 削除したブランチは `anw undo` で作り直せます

**`--worktree`**

- `git worktree remove` で削除します。変更が残っている worktree は削除されません
- `--force` を付けると、変更が残っていても削除します
- 削除した worktree は `anw undo` で戻せません

**`--stash`**

- `git stash drop` で削除します
- 削除した stash は `anw undo` で作り直せます。作り直した stash は、元の前後関係のまま一番上に並びます

### anw tidy

カレントディレクトリ直下のファイルを、更新日時の `yyyy-mm-dd/` ディレクトリに振り分けます。

```sh
anw tidy             # --by date と同じ
anw tidy --by date   # yyyy-mm-dd/ に振り分ける
anw tidy --by month  # yyyy-mm/yyyy-mm-dd/ に振り分ける
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

**`--by month`**

ファイルに加えて、`--by date` で作った `yyyy-mm-dd/` ディレクトリも `yyyy-mm/` の下にまとめます。

```
.
├── 2026-10-27/
│   └── memo.md
└── report.pdf         (2026-10-29 更新)
```

↓

```
.
└── 2026-10/
    ├── 2026-10-27/
    │   └── memo.md
    └── 2026-10-29/
        └── report.pdf
```

- 名前が `yyyy-mm-dd` でないディレクトリは移動しません
- 移動先に同じ日付のディレクトリが既にあれば、中身を1件ずつ移して統合します。同名のファイルには連番を付けます

### anw undo

anw が直前に行った操作を、まとめて元に戻します。

```sh
anw undo
```

| 操作 | 元に戻し方 |
| --- | --- |
| `anw tidy` | 移動したファイルを元の場所に戻す |
| `anw delete --branch` | 削除したブランチを、同じ名前・同じコミットで作り直す。上流ブランチがまだあれば設定し直す |
| `anw delete --stash` | 削除した stash を、同じコミット・同じメッセージで作り直す |

- 戻すのは直前の1回分です
- 先にすべて戻せるかを確かめ、1件でも戻せなければ何もしません（元の場所に同名のファイルがある、同じ名前のブランチがある、コミットがもうない、など）
- `anw kill` で止めたプロセス、`anw edit` での編集、`anw delete --worktree` で削除した worktree は戻せません
- 操作の履歴は、既定では `~/.anw/` に保存されます（[設定](#設定)で変更できます）

### anw run

[設定](#設定) に書いた snippet を実行します。anw のコマンドを、ステップごとに指定したディレクトリで順に実行します。

```sh
anw run --snippet morning
```

- 実行する前に、すべてのステップの dir が存在するかと、anw のコマンドとして正しいかを確認します。1つでも誤りがあれば、何も実行しません
- 途中のステップが失敗するか、一覧の画面でやめると、そこで止めます
- snippet で行ったファイル操作は、`anw undo` でまとめて元に戻せます
- 実行できるのは anw のコマンドだけです。シェルのコマンドや、snippet からの `anw run` は実行できません

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

### snippet

`anw edit --snippet 名前` で書くのがおすすめです。`cd` の行で次からのステップの dir を決め、`anw` の行を1ステップとして、実行する順に書きます。

```sh
# 朝の片付け
cd ~/Desktop
anw tidy
cd ~/Downloads
anw tidy --by month
```

- 空行と `#` で始まる行は読み飛ばします
- `cd` と `anw` 以外の行は書けません。引数を引用符でくくることはできません
- 保存先は `snippets/名前` です。手で置いたファイルも同じように読み込みます

`config.toml` の `[[snippet.名前]]` にも書けます。`snippets/` と同じ名前があれば、その snippet はエラーになります。

```toml
[[snippet.morning]]
dir = "~/Desktop"
args = ["tidy"]

[[snippet.morning]]
dir = "~/Downloads"
args = ["tidy", "--by", "month"]
```

- `dir` はステップを実行するディレクトリです。`~` と環境変数は展開されます
- `args` は anw に渡す引数です。上の例は `cd ~/Desktop && anw tidy && cd ~/Downloads && anw tidy --by month` と同じ動きです

## License

Apache License 2.0
