-- TODO: use innodb/myisam?
create database if not exists bitbot;
use bitbot;

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

create table arbitrage (
    arbitrage_id varchar(100) not null,
    buy_ex varchar(20) not null,
    sell_ex varchar(20) not null,
    pair varchar(10) not null,
    ts timestamp(3) not null,
    buy_price float,
    sell_price float,
    vol float,
    spread float,
    primary key (arbitrage_id)
);

create table order_ack (
    arbitrage_id varchar(100) not null,
    -- depending on the exchanger external_id is either a trade_id or an
    -- order_id. An order Id can be associated with several trade_id.
    external_id varchar(50) not null,
    pair varchar(10) not null,
    exchanger varchar(20),
    side varchar(10)
);
