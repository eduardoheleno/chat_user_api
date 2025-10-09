-- +goose Up
-- +goose StatementBegin

-- nossochat_api.chat_users definition

CREATE TABLE `chat_users` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `chat_id` bigint(20) unsigned NOT NULL,
  `user_id` bigint(20) unsigned NOT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `chat_users_chats_FK` (`chat_id`),
  KEY `chat_users_users_FK` (`user_id`),
  CONSTRAINT `chat_users_chats_FK` FOREIGN KEY (`chat_id`) REFERENCES `chats` (`id`),
  CONSTRAINT `chat_users_users_FK` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=67 DEFAULT CHARSET=utf8;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE chat_users;
-- +goose StatementEnd
