-- MySQL dump 10.16  Distrib 10.1.10-MariaDB, for debian-linux-gnu (x86_64)
--
-- Host: localhost    Database: 
-- ------------------------------------------------------
-- Server version	10.1.10-MariaDB-1~jessie

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8 */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;
--
-- Current Database: `notaryserver`
--

CREATE DATABASE /*!32312 IF NOT EXISTS*/ `notaryserver` /*!40100 DEFAULT CHARACTER SET latin1 */;

USE `notaryserver`;

--
-- Table structure for table `change_category`
--

DROP TABLE IF EXISTS `change_category`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `change_category` (
  `category` varchar(20) NOT NULL,
  PRIMARY KEY (`category`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `change_category`
--

LOCK TABLES `change_category` WRITE;
/*!40000 ALTER TABLE `change_category` DISABLE KEYS */;
INSERT INTO `change_category` VALUES ('deletion'),('update');
/*!40000 ALTER TABLE `change_category` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `changefeed`
--

DROP TABLE IF EXISTS `changefeed`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `changefeed` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `gun` varchar(255) NOT NULL,
  `version` int(11) NOT NULL,
  `sha256` char(64) DEFAULT NULL,
  `category` varchar(20) NOT NULL DEFAULT 'update',
  PRIMARY KEY (`id`),
  KEY `category` (`category`),
  KEY `idx_changefeed_gun` (`gun`),
  CONSTRAINT `changefeed_ibfk_1` FOREIGN KEY (`category`) REFERENCES `change_category` (`category`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `changefeed`
--

LOCK TABLES `changefeed` WRITE;
/*!40000 ALTER TABLE `changefeed` DISABLE KEYS */;
/*!40000 ALTER TABLE `changefeed` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `schema_migrations`
--

DROP TABLE IF EXISTS `schema_migrations`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `schema_migrations` (
  `version` int(11) NOT NULL,
  PRIMARY KEY (`version`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `schema_migrations`
--

LOCK TABLES `schema_migrations` WRITE;
/*!40000 ALTER TABLE `schema_migrations` DISABLE KEYS */;
INSERT INTO `schema_migrations` VALUES (1),(2),(3),(4),(5);
/*!40000 ALTER TABLE `schema_migrations` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `tuf_files`
--

DROP TABLE IF EXISTS `tuf_files`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `tuf_files` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `created_at` timestamp NULL DEFAULT NULL,
  `updated_at` timestamp NULL DEFAULT NULL,
  `deleted_at` timestamp NULL DEFAULT NULL,
  `gun` varchar(255) NOT NULL,
  `role` varchar(255) NOT NULL,
  `version` int(11) NOT NULL,
  `data` longblob NOT NULL,
  `sha256` char(64) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `gun` (`gun`,`role`,`version`),
  KEY `sha256` (`sha256`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `tuf_files`
--

LOCK TABLES `tuf_files` WRITE;
/*!40000 ALTER TABLE `tuf_files` DISABLE KEYS */;
/*!40000 ALTER TABLE `tuf_files` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Current Database: `notarysigner`
--

CREATE DATABASE /*!32312 IF NOT EXISTS*/ `notarysigner` /*!40100 DEFAULT CHARACTER SET latin1 */;

USE `notarysigner`;

--
-- Table structure for table `private_keys`
--

DROP TABLE IF EXISTS `private_keys`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `private_keys` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `created_at` timestamp NULL DEFAULT NULL,
  `updated_at` timestamp NULL DEFAULT NULL,
  `deleted_at` timestamp NULL DEFAULT NULL,
  `key_id` varchar(255) NOT NULL,
  `encryption_alg` varchar(255) NOT NULL,
  `keywrap_alg` varchar(255) NOT NULL,
  `algorithm` varchar(50) NOT NULL,
  `passphrase_alias` varchar(50) NOT NULL,
  `public` blob NOT NULL,
  `private` blob NOT NULL,
  `gun` varchar(255) NOT NULL,
  `role` varchar(255) NOT NULL,
  `last_used` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `key_id` (`key_id`),
  UNIQUE KEY `key_id_2` (`key_id`,`algorithm`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `private_keys`
--

LOCK TABLES `private_keys` WRITE;
/*!40000 ALTER TABLE `private_keys` DISABLE KEYS */;
/*!40000 ALTER TABLE `private_keys` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `schema_migrations`
--

DROP TABLE IF EXISTS `schema_migrations`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `schema_migrations` (
  `version` int(11) NOT NULL,
  PRIMARY KEY (`version`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `schema_migrations`
--

LOCK TABLES `schema_migrations` WRITE;
/*!40000 ALTER TABLE `schema_migrations` DISABLE KEYS */;
INSERT INTO `schema_migrations` VALUES (1),(2);
/*!40000 ALTER TABLE `schema_migrations` ENABLE KEYS */;
UNLOCK TABLES;
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

-- Dump completed on 2017-02-14  6:32:48
