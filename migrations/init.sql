create table if not exists users (
--     id uuid default gen_random_uuid() primary key,
    username varchar(32) primary key,
    password varchar(64) not null,
    coins int default 1000 not null constraint not_negative_check check ( coins >= 0 )
);

create table if not exists items (
--     id uuid default gen_random_uuid() primary key,
    name varchar(32) primary key,
    price int not null constraint not_negative_check check ( price >= 0 )
);

create table if not exists transactions (
    id uuid default gen_random_uuid() primary key,
    time timestamp with time zone default current_timestamp not null,
    fromUser varchar(32) references users(username),
    toUser varchar(32) references users(username) constraint transaction_to_initiator check ( toUser != fromUser ),
    coins int not null constraint not_negative_check check ( coins >= 0 )
);

create table if not exists purchases (
    id uuid default gen_random_uuid() primary key,
    time timestamp with time zone default current_timestamp not null,
    username varchar(32) references users(username),
    item varchar(32) references items(name)
);

-- default items data
insert into items(name, price)
values
    ('t-shirt', 80),
    ('cup', 20),
    ('book', 50),
    ('pen', 10),
    ('powerbank', 200),
    ('hoody', 300),
    ('umbrella', 200),
    ('socks', 10),
    ('wallet', 50),
    ('pink-hoody', 500)

