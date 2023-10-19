CREATE TABLE messages (
      id INT(11) NOT NULL AUTO_INCREMENT,
      username VARCHAR(255) NOT NULL,
      message TEXT NOT NULL,
      time DATETIME NOT NULL default CURRENT_TIMESTAMP,
      color VARCHAR(7) NOT NULL default '#FF00FF',
      PRIMARY KEY (id),
      INDEX idx_time (time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;