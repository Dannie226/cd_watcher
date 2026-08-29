create table versions (
    id bigserial primary key,
    filename text not null,
    foldername text not null

    unique (filename)
);

create table emails (
    event int primary key,
    mainID text,
    chainID text
);
