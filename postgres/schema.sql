CREATE TABLE IF NOT EXISTS public.Vlans
(
    Prefix text COLLATE pg_catalog."default",
    ID integer,
    Demo text COLLATE pg_catalog."default"
);

CREATE TABLE IF NOT EXISTS public.RunningDemos
(
    DemoName text COLLATE pg_catalog."default",
    UserName text COLLATE pg_catalog."default",
    DemoNumber integer,
    Running boolean
);