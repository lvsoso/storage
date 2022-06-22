CREATE TABLE `sync_ops` (  
  `id` bigint  NOT NULL AUTO_INCREMENT,
  `data` JSON NOT  NULL,
  `created_at` TIMESTAMP NOT NULL DEFAULT NOW(),
  PRIMARY KEY (`id`)
)DEFAULT CHARSET=utf8mb4;


CREATE PROCEDURE myops (p1 INT, p2 INT)
BEGIN
  label1: LOOP
    SET p1 = p1 + 1;
    INSERT  INTO  `sync_ops` (`data`) VALUES ( CONCAT('{"op": "', p1 , '", "data": "data"}'));
    IF p1 < p2 THEN    
      ITERATE label1;
    END IF;
    LEAVE label1;
  END LOOP label1;
  SET @x = p1;
END;

CALL myops(0, 1000);