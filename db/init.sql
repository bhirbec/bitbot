-- TODO: use innodb?
create table btc_usd (
	ts timestamp(3) primary key not null,
	orderbooks json
);

create table btc_eur (
	ts timestamp(3) primary key not null,
	orderbooks json
);

create table ltc_btc (
	ts timestamp(3) primary key not null,
	orderbooks json
);
