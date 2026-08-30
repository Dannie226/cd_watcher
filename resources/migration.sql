create table versions (
    id bigserial primary key,
    foldername char(32) not null,

    unique (foldername)
);

create table emails (
    event int primary key,
    mainID text,
    chainID text
);

insert into emails (event) values (1), (2);
