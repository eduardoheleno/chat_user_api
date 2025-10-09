-- +goose Up
-- +goose StatementBegin
CREATE TABLE `contacts` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `sender_id` bigint(20) unsigned NOT NULL,
  `receiver_id` bigint(20) unsigned NOT NULL,
  `status` enum('pending','accepted') DEFAULT 'pending',
  `created_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `unique_pair` (`sender_id`,`receiver_id`),
  KEY `idx_contacts_user1_id` (`sender_id`),
  KEY `idx_contacts_user2_id` (`receiver_id`),
  CONSTRAINT `contacts_users_FK` FOREIGN KEY (`sender_id`) REFERENCES `users` (`id`),
  CONSTRAINT `contacts_users_FK_1` FOREIGN KEY (`receiver_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE contacts;
-- +goose StatementEnd
