create table users (
    id              integer primary key autoincrement,
    email           text,
    token_id        integer
);

create table tokens (
    account       text unique primary key,
    provider      text,
    code          text,
    token         text,
    token_type    text,
    token_refresh text,
    expiry        datetime
);

create table map_google_dropbox (
    google_id  text,
    dropbox_id text
);

create table entries (
    user_id       text,
    title         text unique,
    tag           text,
    priority      text,
    body          text,
    created_at    datetime,
    scheduled     datetime,
    closed        datetime
);

create table sessions (
    sid     text primary key,
    account text
);
