CREATE TABLE `users` (
  `id` int PRIMARY KEY AUTO_INCREMENT,
  `username` varchar(64) NOT NULL,
  `username_unique` varchar(64) NOT NULL,
  `email` varchar(120) NOT NULL,
  `password_hash` varchar(128) NOT NULL,
  `about_me` varchar(140) DEFAULT "",
  `created` datetime NOT NULL,
  `active` tinyint(1) NOT NULL DEFAULT 1,
  `migrate` tinyint(1) NOT NULL DEFAULT 0,
  `country_code` int
);

CREATE TABLE `comments` (
  `id` int PRIMARY KEY AUTO_INCREMENT,
  `parent_id` int,
  `content` text NOT NULL,
  `user_id` int NOT NULL,
  `post_id` int NOT NULL,
  `created` datetime NOT NULL,
  `deleted` tinyint(1) DEFAULT 0
);

CREATE TABLE `comment_votes` (
  `id` int PRIMARY KEY AUTO_INCREMENT,
  `user_id` int NOT NULL,
  `comment_id` int NOT NULL,
  `created` datetime NOT NULL
);

CREATE TABLE `posts` (
  `id` int PRIMARY KEY AUTO_INCREMENT,
  `title` varchar(140) NOT NULL,
  `url_scheme` text,
  `url_base` varchar(253),
  `uri` varchar(255),
  `content` text,
  `created` datetime NOT NULL,
  `user_id` int NOT NULL,
  `deleted` tinyint(1) NOT NULL DEFAULT 0,
  `post_type` tinyint(1) NOT NULL DEFAULT 0
);

CREATE TABLE `votes` (
  `id` int PRIMARY KEY AUTO_INCREMENT,
  `user_id` int NOT NULL,
  `post_id` int NOT NULL,
  `created` datetime NOT NULL
);

ALTER TABLE `comments` ADD FOREIGN KEY (`parent_id`) REFERENCES `comments` (`id`);

ALTER TABLE `comments` ADD FOREIGN KEY (`user_id`) REFERENCES `users` (`id`);

ALTER TABLE `comments` ADD FOREIGN KEY (`post_id`) REFERENCES `posts` (`id`);

ALTER TABLE `comment_votes` ADD FOREIGN KEY (`user_id`) REFERENCES `users` (`id`);

ALTER TABLE `comment_votes` ADD FOREIGN KEY (`comment_id`) REFERENCES `comments` (`id`);

ALTER TABLE `posts` ADD FOREIGN KEY (`user_id`) REFERENCES `users` (`id`);

ALTER TABLE `votes` ADD FOREIGN KEY (`user_id`) REFERENCES `users` (`id`);

ALTER TABLE `votes` ADD FOREIGN KEY (`post_id`) REFERENCES `posts` (`id`);

CREATE UNIQUE INDEX `uc_users_name` ON `users` (`username_unique`);

CREATE UNIQUE INDEX `uc_users_email` ON `users` (`email`);

CREATE INDEX `idx_users` ON `users` (`username`, `email`, `created`);

CREATE INDEX `user_id` ON `comments` (`user_id`);

CREATE INDEX `post_id` ON `comments` (`post_id`);

CREATE INDEX `parent_id` ON `comments` (`parent_id`);

CREATE INDEX `idx_comments` ON `comments` (`user_id`, `post_id`, `created`);

CREATE UNIQUE INDEX `uc_commentvotes` ON `comment_votes` (`user_id`, `comment_id`);

CREATE INDEX `idx_user_id` ON `posts` (`user_id`);

CREATE INDEX `idx_posts` ON `posts` (`user_id`, `created`);

CREATE UNIQUE INDEX `uc_posts_uri` ON `posts` (`uri`);

CREATE UNIQUE INDEX `uc_votes` ON `votes` (`user_id`, `post_id`);