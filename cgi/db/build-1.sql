create table message (id integer primary key autoincrement , title Text, body Text, create_time INTEGER);
create index index_create_time on message (create_time);
