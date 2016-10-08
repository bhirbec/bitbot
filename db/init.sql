-- TODO: use innodb/myisam?

create table orderbooks (
    exchanger varchar(50) not null,
    pair varchar(10) not null,
    ts timestamp(3) not null,
    bids json,
    asks json,
    primary key (exchanger, pair, ts)
);

create table arbitrages (
    buy_ex varchar(20),
    sell_ex varchar(20),
    pair varchar(10) not null,
    ts timestamp(3) not null,
    buy_price float,
    sell_price float,
    vol float,
    spread float,
    key (spread)
);
