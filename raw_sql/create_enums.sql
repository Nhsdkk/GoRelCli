BEGIN;

CREATE TYPE userRole AS ENUM('Admin','User');
CREATE TYPE userType AS ENUM('AdminType','UserType');

COMMIT;