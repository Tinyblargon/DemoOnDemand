CREATE TABLE IF NOT EXISTS public.users
(
    username text COLLATE pg_catalog."default",
    binddn text COLLATE pg_catalog."default"
);

CREATE TABLE IF NOT EXISTS public.vlans
(
    vlanname text COLLATE pg_catalog."default",
    usedbydemo text COLLATE pg_catalog."default"
);

CREATE TABLE IF NOT EXISTS public.runningdemos
(
    demoname text COLLATE pg_catalog."default",
    username text COLLATE pg_catalog."default",
    demonumber integer
);