
CREATE TABLE IF NOT EXISTS `block` (
  `number` bigint NOT NULL,
  `gas_limit` bigint(20) unsigned NOT NULL,
  `gas_used` bigint(20) unsigned NOT NULL,
  `difficulty` bigint NOT NULL,
  `time` bigint(20) unsigned NOT NULL,
  `nonce` bigint(20) unsigned NOT NULL,
  `root` varchar(255) NOT NULL,
  `parent_hash` varchar(255) NOT NULL,
  `tx_hash` varchar(255) NOT NULL,
  `uncle_hash` varchar(255) NOT NULL,
  `extra` blob NOT NULL,
  PRIMARY KEY (`number`)
);

CREATE TABLE IF NOT EXISTS `tx` (
  `hash` varchar(66) NOT NULL,
  `block_number` bigint NOT NULL,
  `nonce` bigint(20) unsigned NOT NULL,
  `to` varchar(42) NOT NULL,
  `from` varchar(42) NOT NULL,
  `value` bigint NOT NULL,
  `data` LONGBLOB NOT NULL,
  PRIMARY KEY (`hash`)
);

CREATE TABLE IF NOT EXISTS `log` (
  `tx_hash` varchar(255) NOT NULL,
  `index` int(10) unsigned NOT NULL,
  `data` blob NOT NULL,
  PRIMARY KEY (`tx_hash`,`index`)
) ;
