SET FOREIGN_KEY_CHECKS=0;

-- ----------------------------
-- Table structure for shorts
-- ----------------------------
DROP TABLE IF EXISTS `shorts`;
CREATE TABLE `shorts` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `uri` varchar(191) NOT NULL COMMENT '''uri 地址''',
  `target` longtext NOT NULL COMMENT '''短链目标''',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uri` (`uri`),
  KEY `idx_shorts_deleted_at` (`deleted_at`)
) ENGINE=InnoDB AUTO_INCREMENT=4 DEFAULT CHARSET=utf8;
SET FOREIGN_KEY_CHECKS=1;
