create table user (
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
    google_id text,
    dropbox_id  text
);

create table events (
    email text,
    due   datetime,
    title text,
    state text,
    body  text
);

create table sessions (
    sid     text primary key,
    account text
);
