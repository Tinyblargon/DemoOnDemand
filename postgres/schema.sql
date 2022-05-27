CREATE TABLE IF NOT EXISTS public.Users
(
    UserName text COLLATE pg_catalog."default",
    BindDn text COLLATE pg_catalog."default"
);

CREATE TABLE IF NOT EXISTS public.Vlans
(
    VlanName text COLLATE pg_catalog."default",
    UsedByDemo text COLLATE pg_catalog."default"
);

CREATE TABLE IF NOT EXISTS public.RunningDemos
(
    DemoName text COLLATE pg_catalog."default",
    UserName text COLLATE pg_catalog."default",
    DemoNumber integer,
    Running boolean
);