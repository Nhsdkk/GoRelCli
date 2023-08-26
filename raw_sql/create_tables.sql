BEGIN;

CREATE TABLE person (
    id SERIAL PRIMARY KEY,
    email text unique not null,
    username text,
    isVerified boolean not null default(false),
    userType userRole not null
);

CREATE TABLE todo (
  id uuid PRIMARY KEY DEFAULT(gen_random_uuid()),
  title text not null ,
  userId int not null ,
  number double precision,
  date timestamptz DEFAULT (now()),
  CONSTRAINT fk_person FOREIGN KEY (userId) references person(id)
);

COMMIT;