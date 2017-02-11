create table users (
    id              integer primary key autoincrement,
    email           text,
    token_id        integer
);

create table tokens (
    account  text unique,
    provider text,
    code     text,
    token    text
);

create table map_google_dropbox (
    google_id  text,
    dropbox_id text
);

create table entries (
    userid        text,
    title         text unique,
    tag           text,
    priority      text,
    body          text,
    create_date   datetime,
    scheduled     datetime,
    closed        datetime
);

create table sessions (
    sid     text primary key,
    account text
);
