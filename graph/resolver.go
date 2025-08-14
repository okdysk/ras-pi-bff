package graph

// THIS CODE WILL BE UPDATED WITH SCHEMA CHANGES. PREVIOUS IMPLEMENTATION FOR SCHEMA CHANGES WILL BE KEPT IN THE COMMENT SECTION. IMPLEMENTATION FOR UNCHANGED SCHEMA WILL BE KEPT.

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/okdysk/ras-pi-bff/db"
	"github.com/okdysk/ras-pi-bff/graph/model"
)

type Resolver struct{}

// スレッドを作成
func (r *mutationResolver) CreateThread(ctx context.Context, title string, content string, author string) (*model.Result, error) {
	tx, err := db.DB.BeginTx(ctx, nil)
	if err != nil {
		msg := "トランザクションの開始に失敗しました"
		return &model.Result{Success: false, Message: &msg, Object: nil}, nil
	}

	// [1] postsにINSERT
	res, err := tx.ExecContext(ctx, `
		INSERT INTO posts (content, author) VALUES (?, ?)
	`, content, author)
	if err != nil {
		tx.Rollback()
		msg := "投稿の作成に失敗しました"
		return &model.Result{Success: false, Message: &msg, Object: nil}, nil
	}

	postID, err := res.LastInsertId()
	if err != nil {
		tx.Rollback()
		msg := "postIDの取得に失敗しました"
		return &model.Result{Success: false, Message: &msg, Object: nil}, nil
	}

	// [2] threadsにINSERT
	res, err = tx.ExecContext(ctx, `
		INSERT INTO threads (title, first_post_id) VALUES (?, ?)
	`, title, postID)
	if err != nil {
		tx.Rollback()
		msg := "スレッドの作成に失敗しました"
		return &model.Result{Success: false, Message: &msg, Object: nil}, nil
	}

	threadID, err := res.LastInsertId()
	if err != nil {
		tx.Rollback()
		msg := "threadIDの取得に失敗しました"
		return &model.Result{Success: false, Message: &msg, Object: nil}, nil
	}

	// [3] postsから取得
	row := tx.QueryRowContext(ctx, `
		SELECT id, content, author, created_at FROM posts WHERE id = ?
	`, postID)

	var post model.Post
	var createdAt time.Time
	if err := row.Scan(&post.ID, &post.Content, &post.Author, &createdAt); err != nil {
		tx.Rollback()
		msg := "投稿取得に失敗しました"
		return &model.Result{Success: false, Message: &msg, Object: nil}, nil
	}
	post.CreatedAt = createdAt.Format(time.RFC3339)

	// [4] COMMIT
	if err := tx.Commit(); err != nil {
		msg := "トランザクションのコミットに失敗しました"
		return &model.Result{Success: false, Message: &msg, Object: nil}, nil
	}

	// [5] Threadを返す
	thread := &model.Thread{
		ID:        fmt.Sprint(threadID),
		Title:     title,
		FirstPost: &post,
	}
	msg := "スレッドを作成しました"
	return &model.Result{Success: true, Message: &msg, Object: thread}, nil
}

// レスポンスを追加
func (r *mutationResolver) AddResponse(ctx context.Context, threadID string, content string, author string) (*model.Result, error) {
	// [1] threadIDをintに変換
	tid, err := strconv.Atoi(threadID)
	if err != nil {
		msg := "threadIDの形式が不正です"
		return &model.Result{Success: false, Message: &msg, Object: nil}, nil
	}

	// [2] トランザクション開始
	tx, err := db.DB.BeginTx(ctx, nil)
	if err != nil {
		msg := "トランザクションの開始に失敗しました"
		return &model.Result{Success: false, Message: &msg, Object: nil}, nil
	}

	// [3] postsにINSERT
	res, err := tx.ExecContext(ctx, `
		INSERT INTO posts (content, author) VALUES (?, ?)
	`, content, author)
	if err != nil {
		tx.Rollback()
		msg := "投稿の作成に失敗しました"
		return &model.Result{Success: false, Message: &msg, Object: nil}, nil
	}

	postID, err := res.LastInsertId()
	if err != nil {
		tx.Rollback()
		msg := "postIDの取得に失敗しました"
		return &model.Result{Success: false, Message: &msg, Object: nil}, nil
	}

	// [4] responsesにINSERT
	res, err = tx.ExecContext(ctx, `
		INSERT INTO responses (thread_id, post_id) VALUES (?, ?)
	`, tid, postID)
	if err != nil {
		tx.Rollback()
		msg := "レスポンスの登録に失敗しました"
		return &model.Result{Success: false, Message: &msg, Object: nil}, nil
	}

	responseID, err := res.LastInsertId()
	if err != nil {
		tx.Rollback()
		msg := "responseIDの取得に失敗しました"
		return &model.Result{Success: false, Message: &msg, Object: nil}, nil
	}

	// [5] postsからSELECT
	row := tx.QueryRowContext(ctx, `
		SELECT id, content, author, created_at FROM posts WHERE id = ?
	`, postID)

	var post model.Post
	var createdAt time.Time
	if err := row.Scan(&post.ID, &post.Content, &post.Author, &createdAt); err != nil {
		tx.Rollback()
		msg := "投稿データの取得に失敗しました"
		return &model.Result{Success: false, Message: &msg, Object: nil}, nil
	}
	post.CreatedAt = createdAt.Format(time.RFC3339)

	// [6] COMMIT
	if err := tx.Commit(); err != nil {
		msg := "トランザクションのコミットに失敗しました"
		return &model.Result{Success: false, Message: &msg, Object: nil}, nil
	}

	// [7] 結果を返す（Responseを含む）
	response := &model.Response{
		ID:   fmt.Sprint(responseID),
		Post: &post,
	}
	msg := "レスポンスを追加しました"
	return &model.Result{
		Success: true,
		Message: &msg,
		Object:  response,
	}, nil
}

// レスポンスを編集
func (r *mutationResolver) UpdatePost(ctx context.Context, postID string, content string, author string) (*model.Result, error) {
	// [1] postIDをintに変換
	id, err := strconv.Atoi(postID)
	if err != nil {
		msg := "postIDの形式が不正です"
		return &model.Result{Success: false, Message: &msg, Object: nil}, nil
	}

	// [2] バリデーション
	if strings.TrimSpace(content) == "" {
		msg := "投稿内容が空です"
		return &model.Result{Success: false, Message: &msg, Object: nil}, nil
	}

	// [3] トランザクション開始
	tx, err := db.DB.BeginTx(ctx, nil)
	if err != nil {
		msg := "トランザクションの開始に失敗しました"
		return &model.Result{Success: false, Message: &msg, Object: nil}, nil
	}

	// [4] postsをUPDATE
	res, err := tx.ExecContext(ctx, `
		UPDATE posts SET content = ?, author = ? WHERE id = ?
	`, content, author, id)
	if err != nil {
		tx.Rollback()
		msg := "投稿の更新に失敗しました"
		return &model.Result{Success: false, Message: &msg, Object: nil}, nil
	}

	// [5] 更新件数チェック
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		tx.Rollback()
		msg := "更新件数の取得に失敗しました"
		return &model.Result{Success: false, Message: &msg, Object: nil}, nil
	}
	if rowsAffected == 0 {
		tx.Rollback()
		msg := "指定された投稿は存在しません"
		return &model.Result{Success: false, Message: &msg, Object: nil}, nil
	}

	// [6] 更新後の投稿をSELECT
	row := tx.QueryRowContext(ctx, `
		SELECT id, content, author, created_at FROM posts WHERE id = ?
	`, id)

	var post model.Post
	var createdAt time.Time
	if err := row.Scan(&post.ID, &post.Content, &post.Author, &createdAt); err != nil {
		tx.Rollback()
		msg := "更新後の投稿取得に失敗しました"
		return &model.Result{Success: false, Message: &msg, Object: nil}, nil
	}
	post.CreatedAt = createdAt.Format(time.RFC3339)

	// [7] COMMIT
	if err := tx.Commit(); err != nil {
		msg := "トランザクションのコミットに失敗しました"
		return &model.Result{Success: false, Message: &msg, Object: nil}, nil
	}

	// [8] 成功結果を返す
	msg := "投稿を更新しました"
	return &model.Result{
		Success: true,
		Message: &msg,
		Object:  &post,
	}, nil
}

// レスポンスを削除
func (r *mutationResolver) DeleteResponse(ctx context.Context, responseID string) (*model.Result, error) {
	// [1] ID変換
	id, err := strconv.Atoi(responseID)
	if err != nil {
		msg := "responseIDの形式が不正です"
		return &model.Result{Success: false, Message: &msg, Object: nil}, nil
	}

	// [2] トランザクション開始
	tx, err := db.DB.BeginTx(ctx, nil)
	if err != nil {
		msg := "トランザクションの開始に失敗しました"
		return &model.Result{Success: false, Message: &msg, Object: nil}, nil
	}

	// [3] post_idを取得
	var postID int
	err = tx.QueryRowContext(ctx, `
		SELECT post_id FROM responses WHERE id = ?
	`, id).Scan(&postID)
	if err != nil {
		tx.Rollback()
		msg := "レスポンスが見つかりません"
		return &model.Result{Success: false, Message: &msg, Object: nil}, nil
	}

	// [4] responses を削除
	res, err := tx.ExecContext(ctx, `
		DELETE FROM responses WHERE id = ?
	`, id)
	if err != nil {
		tx.Rollback()
		msg := "レスポンスの削除に失敗しました"
		return &model.Result{Success: false, Message: &msg, Object: nil}, nil
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil || rowsAffected == 0 {
		tx.Rollback()
		msg := "レスポンスの削除対象が見つかりませんでした"
		return &model.Result{Success: false, Message: &msg, Object: nil}, nil
	}

	// [5] 該当postを削除
	_, err = tx.ExecContext(ctx, `
		DELETE FROM posts WHERE id = ?
	`, postID)
	if err != nil {
		tx.Rollback()
		msg := "レスポンスに紐づく投稿の削除に失敗しました"
		return &model.Result{Success: false, Message: &msg, Object: nil}, nil
	}

	// [6] コミット
	if err := tx.Commit(); err != nil {
		msg := "トランザクションのコミットに失敗しました"
		return &model.Result{Success: false, Message: &msg, Object: nil}, nil
	}

	// [7] 成功返却
	msg := "レスポンスを削除しました"
	return &model.Result{
		Success: true,
		Message: &msg,
		Object: &model.DeleteResult{
			ID: responseID,
		},
	}, nil
}

// スレッドの削除
func (r *mutationResolver) DeleteThread(ctx context.Context, threadID string) (*model.Result, error) {
	// [1] threadIDをintに変換
	id, err := strconv.Atoi(threadID)
	if err != nil {
		msg := "threadIDの形式が不正です"
		return &model.Result{Success: false, Message: &msg, Object: nil}, nil
	}

	// [2] トランザクション開始
	tx, err := db.DB.BeginTx(ctx, nil)
	if err != nil {
		msg := "トランザクションの開始に失敗しました"
		return &model.Result{Success: false, Message: &msg, Object: nil}, nil
	}

	// [3] スレッドのレスポンスに紐づく投稿(post_id)一覧を取得
	rows, err := tx.QueryContext(ctx, `
		SELECT post_id FROM responses WHERE thread_id = ?
	`, id)
	if err != nil {
		tx.Rollback()
		msg := "レスポンスの取得に失敗しました"
		return &model.Result{Success: false, Message: &msg, Object: nil}, nil
	}
	defer rows.Close()

	var postIDs []int
	for rows.Next() {
		var postID int
		if err := rows.Scan(&postID); err != nil {
			tx.Rollback()
			msg := "レスポンスpost_idの読み取りに失敗しました"
			return &model.Result{Success: false, Message: &msg, Object: nil}, nil
		}
		postIDs = append(postIDs, postID)
	}

	// [4] レスに紐づくpostsをバルクDELETE
	if len(postIDs) > 0 {
		// プレースホルダ (?, ?, ...) を作成
		placeholders := make([]string, len(postIDs))
		args := make([]interface{}, len(postIDs))
		for i, postID := range postIDs {
			placeholders[i] = "?"
			args[i] = postID
		}
		query := fmt.Sprintf("DELETE FROM posts WHERE id IN (%s)", strings.Join(placeholders, ","))
		if _, err := tx.ExecContext(ctx, query, args...); err != nil {
			tx.Rollback()
			msg := "レスポンスに紐づく投稿の一括削除に失敗しました"
			return &model.Result{Success: false, Message: &msg, Object: nil}, nil
		}
	}

	// [5] スレッドのfirst_post_idを取得
	var firstPostID int
	err = tx.QueryRowContext(ctx, `
		SELECT first_post_id FROM threads WHERE id = ?
	`, id).Scan(&firstPostID)
	if err != nil {
		tx.Rollback()
		msg := "スレッドのfirst_post_id取得に失敗しました"
		return &model.Result{Success: false, Message: &msg, Object: nil}, nil
	}

	// [6] threadsを削除
	res, err := tx.ExecContext(ctx, `
		DELETE FROM threads WHERE id = ?
	`, id)
	if err != nil {
		tx.Rollback()
		msg := "スレッドの削除に失敗しました"
		return &model.Result{Success: false, Message: &msg, Object: nil}, nil
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		tx.Rollback()
		msg := "削除件数の取得に失敗しました"
		return &model.Result{Success: false, Message: &msg, Object: nil}, nil
	}
	if rowsAffected == 0 {
		tx.Rollback()
		msg := "指定されたスレッドは存在しません"
		return &model.Result{Success: false, Message: &msg, Object: nil}, nil
	}

	// [7] firstPostも削除
	if _, err := tx.ExecContext(ctx, `
		DELETE FROM posts WHERE id = ?
	`, firstPostID); err != nil {
		tx.Rollback()
		msg := "スレッドの最初の投稿の削除に失敗しました"
		return &model.Result{Success: false, Message: &msg, Object: nil}, nil
	}

	// [8] コミット
	if err := tx.Commit(); err != nil {
		msg := "トランザクションのコミットに失敗しました"
		return &model.Result{Success: false, Message: &msg, Object: nil}, nil
	}

	// [9] 成功結果を返す
	msg := "スレッドを削除しました"
	return &model.Result{
		Success: true,
		Message: &msg,
		Object: &model.DeleteResult{
			ID: threadID,
		},
	}, nil
}

// スレッド一覧を取得
func (r *queryResolver) Threads(ctx context.Context) ([]*model.Thread, error) {
	// スレッド一覧＋firstPostをJOINして取得
	rows, err := db.DB.Query(`
		SELECT 
			threads.id, threads.title,
			posts.id, posts.content, posts.author, posts.created_at
		FROM threads
		JOIN posts ON threads.first_post_id = posts.id
	`)
	if err != nil {
		return nil, fmt.Errorf("スレッド取得エラー: %w", err)
	}
	defer rows.Close()

	// スキャンしてスレッド構造体に変換
	var threads []*model.Thread
	for rows.Next() {
		var threadID int
		var title string
		var postID int
		var content, author string
		var createdAt time.Time

		if err := rows.Scan(&threadID, &title, &postID, &content, &author, &createdAt); err != nil {
			return nil, fmt.Errorf("スキャンエラー: %w", err)
		}

		thread := &model.Thread{
			ID:    strconv.Itoa(threadID),
			Title: title,
			FirstPost: &model.Post{
				ID:        strconv.Itoa(postID),
				Content:   content,
				Author:    author,
				CreatedAt: createdAt.Format(time.RFC3339),
			},
		}
		threads = append(threads, thread)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("行ループ中エラー: %w", err)
	}

	return threads, nil
}

// Thread is the resolver for the thread field.
func (r *queryResolver) Thread(ctx context.Context, id string) (*model.Thread, error) {
	threadID, err := strconv.Atoi(id)
	if err != nil {
		return nil, fmt.Errorf("ID変換エラー: %w", err)
	}

	// 対象スレッドの取得
	var title string
	var firstPostID int
	err = db.DB.QueryRow(
		"SELECT title, first_post_id FROM threads WHERE id = ?",
		threadID,
	).Scan(&title, &firstPostID)
	if err == sql.ErrNoRows {
		return nil, nil // not found
	}
	if err != nil {
		return nil, fmt.Errorf("スレッド取得失敗: %w", err)
	}

	// firstPost を取得
	post := &model.Post{}
	var createdAt time.Time
	err = db.DB.QueryRow(
		"SELECT id, content, author, created_at FROM posts WHERE id = ?",
		firstPostID,
	).Scan(&post.ID, &post.Content, &post.Author, &createdAt)
	if err != nil {
		return nil, fmt.Errorf("firstPost取得失敗: %w", err)
	}
	post.CreatedAt = createdAt.Format(time.RFC3339)

	// responses を取得
	rows, err := db.DB.Query(`
		SELECT responses.id, posts.id, posts.content, posts.author, posts.created_at
		FROM responses
		JOIN posts ON responses.post_id = posts.id
		WHERE responses.thread_id = ?
		ORDER BY posts.created_at ASC
	`, threadID)
	if err != nil {
		return nil, fmt.Errorf("レスポンス取得失敗: %w", err)
	}
	defer rows.Close()

	// レスポンスをスキャンして構造体に変換
	var responses []*model.Response
	for rows.Next() {
		var resID, postID int
		var content, author string
		var createdAt time.Time

		if err := rows.Scan(&resID, &postID, &content, &author, &createdAt); err != nil {
			return nil, fmt.Errorf("レスポンススキャン失敗: %w", err)
		}

		responses = append(responses, &model.Response{
			ID: strconv.Itoa(resID),
			Post: &model.Post{
				ID:        strconv.Itoa(postID),
				Content:   content,
				Author:    author,
				CreatedAt: createdAt.Format(time.RFC3339),
			},
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("レスポンスループエラー: %w", err)
	}

	// Thread構造体を構築して返す
	return &model.Thread{
		ID:        id,
		Title:     title,
		FirstPost: post,
		Responses: responses,
	}, nil
}

// スレッドIDでレスポンス部分だけを取得
func (r *queryResolver) Responses(ctx context.Context, threadID string) ([]*model.Response, error) {
	// スレッドIDをintに変換(tid)
	tid, err := strconv.Atoi(threadID)
	if err != nil {
		return nil, fmt.Errorf("スレッドIDが不正です: %w", err)
	}

	// レスポンスを取得
	rows, err := db.DB.Query(`
		SELECT responses.id, posts.id, posts.content, posts.author, posts.created_at
		FROM responses
		JOIN posts ON responses.post_id = posts.id
		WHERE responses.thread_id = ?
		ORDER BY posts.created_at ASC
	`, tid)
	if err != nil {
		return nil, fmt.Errorf("DBクエリエラー: %w", err)
	}
	defer rows.Close()

	// スキャンしてレスポンス構造体に変換
	var responses []*model.Response
	for rows.Next() {
		var resID, postID int
		var content, author string
		var createdAt time.Time

		if err := rows.Scan(&resID, &postID, &content, &author, &createdAt); err != nil {
			return nil, fmt.Errorf("スキャンエラー: %w", err)
		}

		resp := &model.Response{
			ID: strconv.Itoa(resID),
			Post: &model.Post{
				ID:        strconv.Itoa(postID),
				Content:   content,
				Author:    author,
				CreatedAt: createdAt.Format(time.RFC3339),
			},
		}
		responses = append(responses, resp)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("行処理エラー: %w", err)
	}

	return responses, nil
}

// Mutation returns MutationResolver implementation.
func (r *Resolver) Mutation() MutationResolver { return &mutationResolver{r} }

// Query returns QueryResolver implementation.
func (r *Resolver) Query() QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }

// !!! WARNING !!!
// The code below was going to be deleted when updating resolvers. It has been copied here so you have
// one last chance to move it out of harms way if you want. There are two reasons this happens:
//  - When renaming or deleting a resolver the old code will be put in here. You can safely delete
//    it when you're done.
//  - You have helper methods in this file. Move them out to keep these resolver files clean.
/*
	type Resolver struct{}
*/
