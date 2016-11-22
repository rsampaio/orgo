create table user (
    email           text,
    token_id        integer
);

create table tokens (
    id       integer primary key autoincrement,
    provider text,
    account  text unique,
    code     text,
    token    text
);

create table events (
    email text,
    due   datetime,
    title text,
    state text,
    body  text
);
