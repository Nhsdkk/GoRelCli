select *
FROM pg_type
WHERE OID = ANY (select enumtypid
FROM pg_enum);