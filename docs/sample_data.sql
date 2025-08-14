-- DBにサンプルデータを作るSQLスクリプト

-- 既存データがある場合は削除
DROP TABLE IF EXISTS responses;
DROP TABLE IF EXISTS threads;
DROP TABLE IF EXISTS posts;

-- 各テーブルの定義
-- - 投稿テーブル
CREATE TABLE posts (
  id INT PRIMARY KEY AUTO_INCREMENT,
  content TEXT NOT NULL,
  author VARCHAR(100) NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- - スレッドテーブル
CREATE TABLE threads (
  id INT PRIMARY KEY AUTO_INCREMENT,
  title VARCHAR(255) NOT NULL,
  first_post_id INT NOT NULL,
  FOREIGN KEY (first_post_id) REFERENCES posts(id) ON DELETE CASCADE
);

-- - レスポンステーブル
CREATE TABLE responses (
  id SERIAL PRIMARY KEY,
  thread_id INTEGER NOT NULL REFERENCES threads(id) ON DELETE CASCADE,
  post_id INTEGER NOT NULL REFERENCES posts(id) ON DELETE CASCADE
);

-- サンプルデータ挿入
-- - 投稿テーブル
INSERT INTO posts (content, author) VALUES
('スレッド1の1', 'user1'), -- id=1
('スレッド1の2', 'user2'), -- id=2
('スレッド1の3', 'user5'), -- id=3
('スレッド2の1', 'user3'), -- id=4
('スレッド2の2', 'user4'), -- id=5
('スレッド2の3', 'user6'), -- id=6
('スレッド2の4', 'user7'); -- id=7

-- - スレッドテーブル
INSERT INTO threads (title, first_post_id) VALUES
('スレッド1', 1),
('スレッド2', 4);

-- - レスポンステーブル
INSERT INTO responses (thread_id, post_id) VALUES
(1, 2),
(1, 3),
(2, 5),
(2, 6),
(2, 7);