create database if not exists bittrex;
use bittrex;

create table market_summary (
    market_name      varchar(10) not null,
    high             decimal(40, 25) not null,
    low              decimal(40, 25) not null,
    Ask              decimal(40, 25) not null,
    Bid              decimal(40, 25) not null,
    open_buy_orders  int not null,
    open_sell_orders int not null,
    volume           decimal(40, 25) not null,
    last             decimal(40, 25) not null,
    base_volume      decimal(40, 25) not null,
    prev_day         decimal(40, 25) not null,
    timestamp        varchar(30) not null,
    primary key (market_name, timestamp)
);
