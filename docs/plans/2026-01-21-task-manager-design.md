# タスク管理CLI 設計ドキュメント

**作成日**: 2026-01-21
**バージョン**: 1.0

## 概要

個人用タスク管理のためのインタラクティブTUI（Terminal User Interface）アプリケーション。Go言語とBubble Teaフレームワークを使用し、コンポーネントベースのアーキテクチャで実装。GitHub Gistによるクロスデバイス同期機能を提供。

## 1. アーキテクチャ概要

### レイヤー構成

**コンポーネントレイヤー**: 独自の状態と振る舞いをカプセル化した再利用可能なUIコンポーネント
- `List`コンポーネント: スクロール可能なリストで選択、フィルタリング、キーボードナビゲーションをサポート
- `Form`コンポーネント: タスクの作成・編集のための入力フィールドを処理
- `Modal`コンポーネント: 確認ダイアログや詳細ビューのためのオーバーレイ
- `StatusBar`コンポーネント: コンテキスト情報（ショートカット、カウント、フィルタ）を表示
- `HelpModal`コンポーネント: 操作方法を説明するヘルプモーダル

**ビューレイヤー**: コンポーネントを組み合わせてフルスクリーンを構成
- `TasksView`: タスクリスト + ステータスバーを表示するメイン画面
- `KanbanView`: カンバン形式でタスクをステータス別に表示
- `CreateView`: 新規タスク追加用のモーダルフォーム
- `EditView`: 既存タスク編集用のモーダルフォーム
- `FilterView`: フィルタ設定用インターフェース

**状態管理**: Bubble TeaのModelパターンに従った単一のアプリケーション状態ツリー
- `AppModel`: 現在のビューとグローバル状態を管理するルートモデル
- 各ビュー/コンポーネントは`tea.Model`インターフェースを実装する独自のモデルを持つ
- メッセージは上へ、コマンドは下へ流れる

**データレイヤー**: データアクセスのためのリポジトリパターン
- `TaskRepository`インターフェース：ストレージ操作を抽象化
- `SQLiteRepository`実装：データベースクエリを処理
- モデル：バリデーション付きの`Task`、`Category`構造体

## 2. データモデルとデータベーススキーマ

### Taskモデル

```go
type Task struct {
    ID          int64
    Title       string
    Description string
    Status      TaskStatus  // new, working, completed
    Priority    Priority    // low, medium, high
    CategoryID  *int64      // nilの場合はカテゴリなし
    DueDate     *time.Time  // nilの場合は期限なし
    CreatedAt   time.Time
    StartedAt   *time.Time  // workingに変更した時刻
    CompletedAt *time.Time
}
```

### TaskStatus

- `new`: 新規作成されたタスク（未着手）
- `working`: 作業中のタスク
- `completed`: 完了したタスク

### Categoryモデル

```go
type Category struct {
    ID        int64
    Name      string
    Color     string  // TUIでの表示色（例: "blue", "green", "red"）
    CreatedAt time.Time
}
```

### SQLiteスキーマ

```sql
-- カテゴリテーブル
CREATE TABLE categories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    color TEXT NOT NULL,
    created_at DATETIME NOT NULL
);

-- タスクテーブル
CREATE TABLE tasks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    description TEXT,
    status TEXT NOT NULL CHECK(status IN ('new', 'working', 'completed')),
    priority TEXT NOT NULL CHECK(priority IN ('low', 'medium', 'high')),
    category_id INTEGER,
    due_date DATETIME,
    created_at DATETIME NOT NULL,
    started_at DATETIME,
    completed_at DATETIME,
    FOREIGN KEY (category_id) REFERENCES categories(id)
);
```

### Repository操作

- `Create(task *Task)`: 新規タスク作成
- `Update(task *Task)`: タスク更新
- `Delete(id int64)`: タスク削除
- `GetByID(id int64)`: ID指定で取得
- `List(filter Filter)`: フィルタ条件で一覧取得
- `GetCategories()`: カテゴリ一覧取得
- `CreateCategory(category *Category)`: 新規カテゴリ作成

## 3. ユーザーインターフェース

### リストビュー

**表示形式**:
```
[●] [高] タスク名 @カテゴリ 📅 期限
```

- ステータスアイコン: ○(new), ●(working), ✓(completed)
- 優先度の色分け表示

**操作**:
- `j`/`k` (↓/↑): タスク選択移動
- `Enter`: タスク詳細/編集モードへ
- `Space`: ステータス切り替え（new → working → completed）
- `n`: 新規タスク作成
- `d`: 削除確認モーダル表示
- `v`: カンバンビューへ切替
- `f`: フィルタ設定
- `s`: ソート順変更
- `?`/`F1`: ヘルプ表示
- `q`: 終了

### カンバンビュー

**表示形式**:
```
┌─ New (3) ────┬─ Working (2) ─┬─ Completed (5) ─┐
│ [高] タスク1  │ [高] タスク4   │ [中] タスク7     │
│ @開発         │ @開発         │ @開発           │
│ 📅 01/25     │ 📅 01/23      │ ✓ 01/20        │
│              │               │                 │
│ [中] タスク2  │ [低] タスク5   │ ...             │
└──────────────┴───────────────┴─────────────────┘
```

**操作**:
- `h`/`l` (←/→): 列間移動
- `j`/`k` (↓/↑): 同じ列内のタスク選択
- `Enter`: 選択タスクを次の列へ移動（ステータス変更）
- `Shift+Enter`: 前の列へ戻す
- `e`: タスク編集モーダル
- `n`: 新規タスク作成
- `d`: 削除確認
- `v`: リストビューへ切替
- `f`: フィルタ設定
- `?`/`F1`: ヘルプ表示
- `q`: 終了

### タスク作成/編集フォーム

**フィールド**:
- タイトル（必須、1〜200文字）
- 説明（任意、最大1000文字）
- 優先度選択（↑↓で選択）
- カテゴリ入力（テキスト入力 + オートコンプリート）
- 期限日付（YYYY-MM-DD形式）

**操作**:
- `Tab`/`Shift+Tab`: フィールド間移動
- `Enter`: 保存
- `Esc`: キャンセル

**カテゴリ入力の動作**:
- 既存カテゴリをインクリメンタル検索で候補表示
- 存在しないカテゴリ名を入力した場合、Enter時に確認モーダル表示
- 「カテゴリ「〇〇」を新規作成しますか？ 色を選択: [青/緑/赤/黄...]」
- Yes選択で新規カテゴリ作成 + タスクに紐付け

### ヘルプモーダル

**リストビュー**:
```
┌─ ヘルプ - リストビュー ─────────────────┐
│                                        │
│ 移動:                                  │
│   j/↓      : 下へ移動                  │
│   k/↑      : 上へ移動                  │
│                                        │
│ タスク操作:                            │
│   Enter    : タスク詳細/編集           │
│   Space    : ステータス切替            │
│   n        : 新規タスク作成            │
│   d        : タスク削除                │
│                                        │
│ 表示:                                  │
│   v        : カンバンビューへ切替      │
│   f        : フィルタ設定              │
│   ?/F1     : このヘルプ                │
│   q        : 終了                      │
│                                        │
│         [Esc または ? で閉じる]        │
└────────────────────────────────────────┘
```

**カンバンビュー**:
```
┌─ ヘルプ - カンバンビュー ───────────────┐
│                                        │
│ 移動:                                  │
│   h/←      : 左の列へ                  │
│   l/→      : 右の列へ                  │
│   j/↓      : 列内で下へ                │
│   k/↑      : 列内で上へ                │
│                                        │
│ タスク操作:                            │
│   Enter    : 次のステータスへ移動      │
│   Shift+Enter: 前のステータスへ戻す    │
│   e        : タスク編集                │
│   n        : 新規タスク作成            │
│   d        : タスク削除                │
│                                        │
│ 表示:                                  │
│   v        : リストビューへ切替        │
│   f        : フィルタ設定              │
│   ?/F1     : このヘルプ                │
│   q        : 終了                      │
│                                        │
│         [Esc または ? で閉じる]        │
└────────────────────────────────────────┘
```

### ステータスバー

画面下部に常に簡易ヘルプと状態を表示:
```
[?]ヘルプ [n]新規 [v]ビュー切替 [f]フィルタ [q]終了 [☁ 同期済み 2分前]
```

## 4. 状態管理とメッセージフロー

### AppModel（ルートモデル）

```go
type AppModel struct {
    currentView  View  // tasks_list, tasks_kanban, create_task, edit_task, filter, help
    tasksList    *TasksListModel
    tasksKanban  *TasksKanbanModel
    createForm   *CreateFormModel
    editForm     *EditFormModel
    filterView   *FilterModel
    helpModal    *HelpModalModel
    repository   TaskRepository
    width, height int
}
```

### メッセージフロー

1. **キー入力** → `Update()`メソッドでメッセージ処理
   - グローバルキー（?、q、v）: AppModelで処理
   - ビュー固有キー: 各ビューのモデルに委譲

2. **状態遷移のメッセージ**:
   - `taskCreatedMsg`: タスク作成完了 → リスト再読み込み
   - `taskUpdatedMsg`: タスク更新完了 → 表示更新
   - `taskDeletedMsg`: タスク削除完了 → リストから削除
   - `viewSwitchMsg`: ビュー切替要求
   - `categoryCreateConfirmMsg`: カテゴリ新規作成確認
   - `filterAppliedMsg`: フィルタ適用完了

3. **データ更新フロー**:
   ```
   ユーザー操作 → Message生成 → Update()
   → Repository操作（DB更新） → Command実行
   → 完了Message → View更新
   ```

4. **モーダル管理**:
   - モーダル表示中は背景ビューの入力を無効化
   - Escキーでモーダルを閉じて元のビューに戻る
   - モーダルスタック管理で多重モーダルにも対応

## 5. フィルタリングとソート機能

### Filterモデル

```go
type Filter struct {
    Status     []TaskStatus  // 複数選択可能
    Priority   []Priority    // 複数選択可能
    Categories []int64       // カテゴリID、複数選択可能
    DateRange  DateRange     // today, this_week, overdue, all
    SearchText string        // タイトル・説明の部分一致検索
}

type DateRange int
const (
    AllDates DateRange = iota
    Today
    ThisWeek
    Overdue
    NoDueDate
)
```

### FilterView UI

```
┌─ フィルタ設定 ─────────────────┐
│ ステータス:                    │
│  [✓] New                       │
│  [✓] Working                   │
│  [ ] Completed                 │
│                                │
│ 優先度:                        │
│  [✓] 高   [✓] 中   [ ] 低     │
│                                │
│ カテゴリ:                      │
│  [✓] 開発                      │
│  [ ] デザイン                  │
│  [✓] ドキュメント              │
│                                │
│ 期限:                          │
│  ( ) すべて                    │
│  (●) 今週                      │
│  ( ) 期限切れのみ              │
│                                │
│ 検索: [_______________]        │
│                                │
│     [適用]    [クリア]         │
└────────────────────────────────┘
```

### ソート機能

`s`キーでソート順変更メニュー表示:
- 作成日時（新しい順/古い順）
- 期限日（近い順/遠い順）
- 優先度（高→低/低→高）
- ステータス順
- カテゴリ名

### 適用動作

- フィルタ/ソート適用後、リストビュー・カンバンビュー両方に反映
- ステータスバーに現在のフィルタ条件を表示（例: `[フィルタ: 開発, 高優先度]`）
- フィルタクリアで全件表示に戻る

## 6. プロジェクト構造とパッケージ設計

### ディレクトリ構造

```
task-management/
├── cmd/
│   └── task/
│       └── main.go              // エントリーポイント
├── internal/
│   ├── app/
│   │   ├── app.go               // AppModel、メインループ
│   │   └── messages.go          // カスタムメッセージ定義
│   ├── ui/
│   │   ├── components/
│   │   │   ├── list.go          // Listコンポーネント
│   │   │   ├── form.go          // Formコンポーネント
│   │   │   ├── modal.go         // Modalコンポーネント
│   │   │   ├── statusbar.go     // StatusBarコンポーネント
│   │   │   └── help.go          // HelpModalコンポーネント
│   │   ├── views/
│   │   │   ├── tasks_list.go    // リストビュー
│   │   │   ├── tasks_kanban.go  // カンバンビュー
│   │   │   ├── filter.go        // フィルタビュー
│   │   │   └── task_form.go     // タスク作成/編集ビュー
│   │   └── styles/
│   │       └── styles.go         // 色、スタイル定義（lipgloss）
│   ├── domain/
│   │   ├── task.go              // Task、Category モデル
│   │   ├── filter.go            // Filter モデル
│   │   └── repository.go        // TaskRepository インターフェース
│   ├── repository/
│   │   ├── sqlite.go            // SQLiteRepository 実装
│   │   └── migrations.go        // DBマイグレーション
│   └── sync/
│       ├── gist.go              // GitHub Gist同期機能
│       └── exporter.go          // JSON エクスポート/インポート
├── docs/
│   └── plans/
│       └── 2026-01-21-task-manager-design.md
├── go.mod
└── go.sum
```

### 主要な依存パッケージ

- `github.com/charmbracelet/bubbletea`: TUIフレームワーク
- `github.com/charmbracelet/lipgloss`: スタイリング
- `github.com/charmbracelet/bubbles`: 再利用可能なコンポーネント
- `modernc.org/sqlite`: Pure Go SQLiteドライバー（cgo不要）
- `github.com/google/go-github/v58/github`: GitHub API クライアント
- `golang.org/x/oauth2`: OAuth認証

### 設計原則

- `internal/domain`: ビジネスロジックとモデル、UIやDBに依存しない
- `internal/repository`: データアクセス層、インターフェースで抽象化
- `internal/ui`: UI層、domain層に依存するがrepository層には直接依存しない
- `internal/app`: 各層を統合、依存性注入
- `internal/sync`: 同期機能、domainとrepositoryに依存

## 7. エラーハンドリングとバリデーション

### エラー表示方法

- UI内にエラー通知コンポーネント（トースト）を追加
- 画面上部に一時的な通知として表示
- 数秒後に自動的に消えるが、Escキーで即座に閉じることも可能

```
┌────────────────────────────────────┐
│ ⚠ エラー: タスクの保存に失敗しました │
└────────────────────────────────────┘
```

### バリデーション

**タスク作成/編集時**:
- タイトル: 必須、1〜200文字
- 説明: 任意、最大1000文字
- 期限日: 有効な日付形式（YYYY-MM-DD）
- カテゴリ: 存在確認、または新規作成確認
- バリデーションエラーはフォーム内にインラインで表示

**データベースエラー**:
- 接続エラー: 起動時にエラーメッセージ表示して終了
- 書き込みエラー: トースト通知で表示、操作をロールバック
- 外部キー制約違反: ユーザーフレンドリーなメッセージに変換

**リカバリー戦略**:
- DB操作はトランザクション内で実行
- 失敗時は前の状態に戻す
- 楽観的UI更新（先にUIを更新、失敗したら戻す）で快適な操作感

**ログ出力**:
- デバッグ用に`~/.task-management/debug.log`にログ出力
- エラー発生時の詳細情報を記録
- 本番では無効化可能

## 8. テスト戦略

### テストレイヤー

**ドメイン層のユニットテスト**:
- `Task`、`Category`のバリデーションロジック
- `Filter`の条件判定ロジック
- テーブル駆動テストで様々なケースをカバー

**リポジトリ層のテスト**:
- インメモリSQLite（`:memory:`）を使用
- CRUD操作の正確性を検証
- フィルタリング、ソート機能のテスト
- トランザクションのロールバック動作確認

**UI層のテスト**:
- Bubble Teaのモデル更新ロジックをテスト
- メッセージ送信→Update()→状態変化を検証
- 実際の描画はテストせず、状態遷移のみテスト
- モックRepositoryを使用してDB依存を排除

**同期機能のテスト**:
- モックGist APIを使用
- JSON エクスポート/インポートの正確性
- マージロジックのテスト（競合解決など）

### テストファイル配置

```
internal/
├── domain/
│   ├── task.go
│   └── task_test.go
├── repository/
│   ├── sqlite.go
│   └── sqlite_test.go
├── sync/
│   ├── gist.go
│   └── gist_test.go
└── ui/
    ├── views/
    │   ├── tasks_list.go
    │   └── tasks_list_test.go
```

### テスト実行

- `go test ./...`: 全テスト実行
- `go test -race ./...`: 競合検出付き
- `go test -cover ./...`: カバレッジ測定

### 手動テスト

- 各ビューの操作フローをチェックリスト化
- キーボードショートカットの動作確認
- 画面サイズ変更への対応確認

## 9. ビルド、配布、初回セットアップ

### ビルド設定

Makefileで簡単ビルド:
```makefile
build:
	go build -o bin/task ./cmd/task

install:
	go install ./cmd/task

test:
	go test -v -race -cover ./...

run:
	go run ./cmd/task
```

### 初回起動時の動作

1. `~/.task-management/`ディレクトリが存在しない場合は自動作成
2. `~/.task-management/tasks.db`にSQLiteデータベース作成
3. マイグレーション実行（テーブル作成）
4. デフォルトカテゴリを3つ作成（例: "仕事"、"個人"、"その他"）
5. ウェルカムメッセージをモーダル表示:

```
┌─ Welcome to Task Manager ──────┐
│                                │
│ タスク管理へようこそ！          │
│                                │
│ 使い方:                        │
│ - n: 新しいタスクを作成        │
│ - ?: ヘルプを表示              │
│ - v: ビューを切り替え          │
│                                │
│ データの保存場所:              │
│ ~/.task-management/tasks.db    │
│                                │
│      [Enter で開始]            │
└────────────────────────────────┘
```

### 配布方法

- GitHubリリースでバイナリ配布（Linux、macOS、Windows）
- `go install`でインストール可能
- Homebrewタップも検討可能（将来的に）

### 設定ファイル

`~/.task-management/config.yaml`で設定可能:
```yaml
# 色テーマ
theme: default  # default, dark, light

# デフォルトビュー
default_view: list  # list, kanban

# 同期設定
sync:
  enabled: true
  auto_pull_on_start: true
  auto_push_on_change: false
  gist_id: ""

# GitHub Personal Access Token（環境変数推奨）
# github_token: "ghp_xxxxx"
```

設定ファイルがない場合はデフォルト値を使用。

### データバックアップ

- `tasks.db`ファイルをコピーするだけでバックアップ完了
- GitHub Gist同期により自動的にクラウドバックアップ

## 10. GitHub Gist同期機能

### JSONエクスポート形式

```json
{
  "version": "1.0",
  "exported_at": "2026-01-21T12:34:56Z",
  "categories": [
    {
      "id": 1,
      "name": "仕事",
      "color": "blue",
      "created_at": "2026-01-20T10:00:00Z"
    }
  ],
  "tasks": [
    {
      "id": 1,
      "title": "設計書を書く",
      "description": "APIの設計書を作成",
      "status": "working",
      "priority": "high",
      "category_id": 1,
      "due_date": "2026-01-25T00:00:00Z",
      "created_at": "2026-01-20T10:30:00Z",
      "started_at": "2026-01-21T09:00:00Z",
      "completed_at": null
    }
  ]
}
```

### 同期機能の実装

**初期設定**:
1. `g`キーまたは`sync`コマンド: 同期設定画面表示
2. GitHub Personal Access Token入力（Gist権限が必要）
3. `~/.task-management/config.yaml`に保存（環境変数推奨）
4. 初回は新規Gist作成、以降は同じGist IDを使用

**プッシュ（ローカル→Gist）**:
1. SQLiteからJSON形式にエクスポート
2. GitHub Gist APIでGistを更新
3. 最終同期時刻を記録

**プル（Gist→ローカル）**:
1. GistからJSON取得
2. ローカルDBとマージ:
   - **Last-Write-Wins**: タイムスタンプで新しい方を採用
   - 競合検出: 同じタスクが両方で変更されている場合
   - 競合解決UI: ユーザーに選択させる

**自動同期**:
- 起動時に自動プル（設定で無効化可能）
- タスク作成/更新/削除時に自動プッシュ（設定で無効化可能）
- または手動同期のみ（`Ctrl+S`で同期）

### TUI操作

**同期設定画面（初回）**:
```
┌─ GitHub同期設定 ───────────────────┐
│                                    │
│ Personal Access Token:             │
│ [ghp_xxxxxxxxxxxxxxxxxxxxx______]  │
│                                    │
│ 権限: 'gist' が必要です            │
│                                    │
│ トークンの取得方法:                │
│ github.com/settings/tokens/new     │
│                                    │
│     [保存して同期開始]             │
└────────────────────────────────────┘
```

**同期状態表示（ステータスバー）**:
- `[☁ 同期済み 2分前]`
- `[⚠ 未同期の変更あり]`
- `[↓ 同期中...]`

**同期コマンド**:
- `Ctrl+S`: 今すぐ同期
- `g`: 同期設定

### セキュリティ考慮

- Personal Access Tokenは環境変数`TASK_GITHUB_TOKEN`または暗号化して保存
- プライベートGistを使用（デフォルト）
- 通信はHTTPS
- トークンは最小権限（gistのみ）

### 競合解決UI

同じタスクが両方で変更されている場合:
```
┌─ 競合の解決 ───────────────────────┐
│                                    │
│ タスク「設計書を書く」で競合       │
│                                    │
│ ローカル:                          │
│ - ステータス: working              │
│ - 更新: 2026-01-21 10:00           │
│                                    │
│ リモート:                          │
│ - ステータス: completed            │
│ - 更新: 2026-01-21 10:05           │
│                                    │
│ どちらを採用しますか？             │
│                                    │
│  [ローカル]  [リモート]  [両方保持]│
└────────────────────────────────────┘
```

## 実装の優先順位

### Phase 1: コア機能
1. プロジェクト構造セットアップ
2. データモデルとSQLiteリポジトリ
3. 基本的なリストビュー（タスク一覧、作成、削除）
4. ステータス切り替え機能

### Phase 2: UI拡張
5. カンバンビュー
6. フィルタリング機能
7. ソート機能
8. ヘルプモーダル

### Phase 3: 高度な機能
9. カテゴリ管理
10. エラーハンドリングとバリデーション
11. テスト実装

### Phase 4: 同期機能
12. JSON エクスポート/インポート
13. GitHub Gist同期
14. 競合解決UI

## まとめ

本設計では、Go言語とBubble Teaを使用した個人用タスク管理TUIアプリケーションを定義しました。コンポーネントベースのアーキテクチャにより、拡張性とメンテナンス性を確保しています。

主な特徴:
- インタラクティブなTUI（リストビュー、カンバンビュー）
- 3段階のタスクステータス（new, working, completed）
- カテゴリ、優先度、期限による管理
- 強力なフィルタリングとソート機能
- GitHub Gistによるクロスデバイス同期
- 包括的なテスト戦略

この設計に基づいて実装を進めることで、使いやすく信頼性の高いタスク管理ツールが完成します。
