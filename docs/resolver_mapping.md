# リゾルバ設計仕様書

## Query

| フィールド | 説明 | Go関数 | 対応するDB処理（概要） |
|------------|------|--------|--------------------------|
| threads | 全スレッドを一覧取得 | `(*queryResolver) Threads(ctx)` | `SELECT * FROM threads` |
| thread(id: ID!) | スレッドIDを指定して対象スレッドを取得 | `(*queryResolver) Thread(ctx, id)` | `SELECT * FROM threads WHERE id = ?` |
| responses(threadId: ID!) | スレッドIDを指定して、該当スレッドに属するレスポンス一覧を投稿日時順で取得 | `(*queryResolver) Responses(ctx, threadId)` | `SELECT posts.* FROM responses JOIN posts ON responses.post_id = posts.id WHERE responses.thread_id = ? ORDER BY posts.created_at ASC` |

## Thread

| フィールド | 説明 | Go関数 | 対応するDB処理（概要） |
|------------|------|--------|--------------------------|
| id | スレッドID | 自動解決 | - |
| title | スレッドタイトル | 自動解決 | - |
| firstPost | スレッドの最初の投稿を取得（threadsテーブル内のfirst_post_idを参照） | `(*threadResolver) FirstPost(ctx, obj)` | `SELECT * FROM posts WHERE id = first_post_id` |
| responses | スレッドに属するレスポンス一覧 | `(*threadResolver) Responses(ctx, obj)` | `SELECT responses.* FROM responses WHERE thread_id = ?` |

## Response

| フィールド | 説明 | Go関数 | 対応するDB処理（概要） |
|------------|------|--------|--------------------------|
| id | レスID | 自動解決 | - |
| post | レスの投稿情報（post_idでpostsを参照） | `(*responseResolver) Post(ctx, obj)` | `SELECT * FROM posts WHERE id = ?` |

## Post

| フィールド | 説明 | Go関数 | 備考 |
|------------|------|--------|------|
| id | 投稿ID | 自動解決 | - |
| content | 投稿本文 | 自動解決 | - |
| author | 投稿者名 | 自動解決 | - |
| createdAt | 投稿日時 | 自動解決 | - |

---

## Mutation

| フィールド | 説明 | Go関数 | 対応するDB処理（概要） |
|------------|------|--------|--------------------------|
| createThread(title, content, author) | スレッド作成（最初の投稿付き） | `(*mutationResolver) CreateThread(ctx, title, content, author)` | postsにINSERT → threadsにINSERT。成功時にThreadをobjectとして返却 |
| addResponse(threadId, content, author) | スレッドにレスを追加 | `(*mutationResolver) AddResponse(ctx, threadId, content, author)` | postsにINSERT → responsesにINSERT。Responseをobjectとして返却 |
| updatePost(postId, content, author) | 投稿の内容と投稿者を更新 | `(*mutationResolver) UpdatePost(ctx, postId, content, author)` | postsをUPDATE。成功時に更新後のPostを返却 |
| deleteThread(threadId) | スレッドを削除（該当Post/Responseも含めて） | `(*mutationResolver) DeleteThread(ctx, threadId)` | responses → posts → threads の順でDELETE。成功時はDeleteResultを返却 |
| deleteResponse(responseId) | レスを削除（該当Postも削除） | `(*mutationResolver) DeleteResponse(ctx, responseId)` | responsesをDELETE → postsをDELETE。成功時はDeleteResultを返却 |
---

## 補足メモ

- `responses` フィールドは、N+1問題を回避するため、`posts` テーブルとのJOINを前提としています。
- `firstPost` フィールドは、`threads.first_post_id` を参照して該当投稿を特定します。
- 各IDの意味（Thread.id, Post.id など）が混同されないよう、引数名や説明文に明記しています。
- すべてのmutationは `Result` 型を返すよう統一され、成功可否・メッセージ・返却オブジェクトを一貫して扱えるようになっている。
- `Result.object` はUnion型 `ResultObject = Thread | Response | Post | DeleteResult` として定義。
- 削除処理では `DeleteResult` 型（IDのみ返却）をobjectに格納する。