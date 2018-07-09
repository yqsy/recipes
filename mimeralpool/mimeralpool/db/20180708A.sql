CREATE DATABASE mimeralpool;

USE mimeralpool;


CREATE TABLE b_currency (
  id CHAR(32),
  name VARCHAR(32),
  remain DECIMAL(23,8),
  today_remain DECIMAL(23,8),
  PRIMARY KEY (id)
);

CREATE TABLE b_user_computing_power (
  id CHAR(32),
  currency_id CHAR(32),
  computing_power INT,
  PRIMARY KEY (id, currency_id)
);

CREATE TABLE f_use_computing_power (
  id CHAR(64),
  user_computing_power_id CHAR(32),
  amount INT,
  power_change_mode INT,
  PRIMARY KEY (id)
);


CREATE TABLE b_currency_sponsor (
  currency_id  CHAR(32),
  sponsor_id CHAR(32),
  amount DECIMAL(23,8)
);

CREATE TABLE b_sponsor (
  id CHAR(32),
  name VARCHAR(64),
  PRIMARY KEY (id)
);


CREATE TABLE d_dict_type (
  id int AUTO_INCREMENT,
  code VARCHAR(64),
  comment  VARCHAR(64),
  PRIMARY KEY (id)
);

CREATE TABLE d_dict_item (
  id int AUTO_INCREMENT,
  type_id int,
  value INT,
  comment VARCHAR (64),
  PRIMARY KEY (id)
);



INSERT d_dict_type SET code="power_change_mode", comment="算力变动原因";
INSERT d_dict_item SET type_id=1,value=0,comment="入币增加算力";
INSERT d_dict_item SET type_id=1,value=1,comment="邀请好友增加算力";
INSERT d_dict_item SET type_id=1,value=2,comment="注销清空算力"
