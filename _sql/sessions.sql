SET NAMES utf8;
SET time_zone = '+00:00';
SET foreign_key_checks = 0;
SET sql_mode = 'NO_AUTO_VALUE_ON_ZERO';

DROP TABLE IF EXISTS `sessions`;
CREATE TABLE `sessions` (
                         `id`       varchar(255) NOT NULL,
                         `userid`   varchar(255) NOT NULL,
                         `username` varchar(255) NOT NULL,
                         `expires`  datetime     NOT NULL,
                         PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
