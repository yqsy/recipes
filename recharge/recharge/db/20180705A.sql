CREATE DATABASE recharge;

USE recharge;

CREATE TABLE s_user (
  id CHAR(32),
  phone CHAR(11),
  email VARCHAR(255),
  invite_code VARCHAR(255),
  passwd VARCHAR(60),
  balance DECIMAL(18,2),
  PRIMARY KEY (id),
  UNIQUE KEY (phone),
  UNIQUE KEY (email)
);


CREATE TABLE f_goldin_flow (
  id CHAR(64),
  user_id CHAR(32),
  amount DECIMAL(18,2),
  PRIMARY KEY (id),
  INDEX (user_id)
);


