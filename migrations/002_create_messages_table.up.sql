CREATE TABLE messages (
      id INT(11) NOT NULL AUTO_INCREMENT,
      username VARCHAR(255) NOT NULL,
      message TEXT NOT NULL,
      time INT(11) NOT NULL,
      PRIMARY KEY (id)
);