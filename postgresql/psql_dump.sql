SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

SET default_tablespace = '';

SET default_table_access_method = heap;

CREATE TABLE public.messages (
    id integer NOT NULL,
    "from" text NOT NULL,
    "to" text NOT NULL,
    msg text NOT NULL
);

ALTER TABLE public.messages OWNER TO postgres;

CREATE SEQUENCE public.messages_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER TABLE public.messages_id_seq OWNER TO postgres;

ALTER SEQUENCE public.messages_id_seq OWNED BY public.messages.id;

CREATE TABLE public.users (
    login text NOT NULL,
    password text NOT NULL
);

ALTER TABLE public.users OWNER TO postgres;

ALTER TABLE ONLY public.messages ALTER COLUMN id SET DEFAULT nextval('public.messages_id_seq'::regclass);

SELECT pg_catalog.setval('public.messages_id_seq', 7, true);

ALTER TABLE ONLY public.messages
    ADD CONSTRAINT messages_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (login);

ALTER TABLE ONLY public.messages
    ADD CONSTRAINT messages_from_fkey FOREIGN KEY ("from") REFERENCES public.users(login);

ALTER TABLE ONLY public.messages
    ADD CONSTRAINT messages_to_fkey FOREIGN KEY ("to") REFERENCES public.users(login);
