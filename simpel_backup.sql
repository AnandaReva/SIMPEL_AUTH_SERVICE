--
-- PostgreSQL database dump
--

-- Dumped from database version 17.2
-- Dumped by pg_dump version 17.2

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET transaction_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: device; Type: SCHEMA; Schema: -; Owner: revaananda
--

CREATE SCHEMA device;


ALTER SCHEMA device OWNER TO revaananda;

--
-- Name: sysuser; Type: SCHEMA; Schema: -; Owner: revaananda
--

CREATE SCHEMA sysuser;


ALTER SCHEMA sysuser OWNER TO revaananda;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: data; Type: TABLE; Schema: device; Owner: revaananda
--

CREATE TABLE device.data (
    id bigint NOT NULL,
    unit_id bigint NOT NULL,
    tstamp bigint DEFAULT (EXTRACT(epoch FROM now()))::bigint NOT NULL,
    voltage double precision NOT NULL,
    current double precision NOT NULL,
    power double precision NOT NULL,
    energy double precision NOT NULL,
    frequency double precision NOT NULL,
    power_factor double precision NOT NULL
);


ALTER TABLE device.data OWNER TO revaananda;

--
-- Name: data_id_seq; Type: SEQUENCE; Schema: device; Owner: revaananda
--

CREATE SEQUENCE device.data_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE device.data_id_seq OWNER TO revaananda;

--
-- Name: data_id_seq; Type: SEQUENCE OWNED BY; Schema: device; Owner: revaananda
--

ALTER SEQUENCE device.data_id_seq OWNED BY device.data.id;


--
-- Name: device_id_sq; Type: SEQUENCE; Schema: public; Owner: revaananda
--

CREATE SEQUENCE public.device_id_sq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.device_id_sq OWNER TO revaananda;

--
-- Name: unit; Type: TABLE; Schema: device; Owner: revaananda
--

CREATE TABLE device.unit (
    id bigint DEFAULT nextval('public.device_id_sq'::regclass) NOT NULL,
    name character varying(255) NOT NULL,
    status integer NOT NULL,
    salt character varying(64) NOT NULL,
    salted_password character varying(128) NOT NULL,
    data jsonb NOT NULL,
    create_tstamp bigint DEFAULT (EXTRACT(epoch FROM now()))::bigint
);


ALTER TABLE device.unit OWNER TO revaananda;

--
-- Name: session; Type: TABLE; Schema: sysuser; Owner: revaananda
--

CREATE TABLE sysuser.session (
    session_id character varying(16) NOT NULL,
    user_id bigint NOT NULL,
    session_hash character varying(128) NOT NULL,
    tstamp bigint NOT NULL,
    st integer NOT NULL,
    last_ms_tstamp bigint,
    last_sequence bigint
);


ALTER TABLE sysuser.session OWNER TO revaananda;

--
-- Name: token; Type: TABLE; Schema: sysuser; Owner: revaananda
--

CREATE TABLE sysuser.token (
    user_id bigint NOT NULL,
    token character varying(128) NOT NULL,
    tstamp bigint NOT NULL
);


ALTER TABLE sysuser.token OWNER TO revaananda;

--
-- Name: user_id_seq; Type: SEQUENCE; Schema: sysuser; Owner: revaananda
--

CREATE SEQUENCE sysuser.user_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE sysuser.user_id_seq OWNER TO revaananda;

--
-- Name: user; Type: TABLE; Schema: sysuser; Owner: revaananda
--

CREATE TABLE sysuser."user" (
    username character varying(30) NOT NULL,
    full_name character varying(128) NOT NULL,
    st integer NOT NULL,
    salt character varying(64) NOT NULL,
    saltedpassword character varying(128) NOT NULL,
    data jsonb NOT NULL,
    id bigint DEFAULT nextval('sysuser.user_id_seq'::regclass) NOT NULL,
    role character varying(128) NOT NULL
);


ALTER TABLE sysuser."user" OWNER TO revaananda;

--
-- Name: data id; Type: DEFAULT; Schema: device; Owner: revaananda
--

ALTER TABLE ONLY device.data ALTER COLUMN id SET DEFAULT nextval('device.data_id_seq'::regclass);


--
-- Data for Name: data; Type: TABLE DATA; Schema: device; Owner: revaananda
--

COPY device.data (id, unit_id, tstamp, voltage, current, power, energy, frequency, power_factor) FROM stdin;
\.


--
-- Data for Name: unit; Type: TABLE DATA; Schema: device; Owner: revaananda
--

COPY device.unit (id, name, status, salt, salted_password, data, create_tstamp) FROM stdin;
1	device_i	1	d8fkwR4xiW5lpxYS	90ca22705c91a3e3daca6d6330d09d4065eb8948d7defdac0276fef3f1f963e8	{}	1739370518
2	device_j	1	elRaD507UDQEIky9	c20b232f343ddef833bfe2b0839f051c2d7cf9d6c438abacbf7e8adaf184d663	{}	1739373528
\.


--
-- Data for Name: session; Type: TABLE DATA; Schema: sysuser; Owner: revaananda
--

COPY sysuser.session (session_id, user_id, session_hash, tstamp, st, last_ms_tstamp, last_sequence) FROM stdin;
npGnYG2IUWnDauIa	1	ad8c333e1141e78d4d87ff2f3e55d80529094fe6f21685c054557f3beb2e39fa	1738681697	1	\N	\N
Vp9x2QkGfNl20r3P	9	2ed6d9f9c3d3d6267df55ca5dc14ebf676f0b9b729662324e2a1c99c203275c4	1738725661	1	\N	\N
2wsgprZ2J80vDcfp	12	fe38f2559bf1962363fcf6620a751c0101b7446f91e91831a4342299f4494cb5	1738944883	1	\N	\N
3h0mbshYGSUS9Ls6	13	6fcbf8380ab3608ad288f889b98bca5455ab781810e89fdd92ca266fcec2a4a0	1739688665	1	\N	\N
\.


--
-- Data for Name: token; Type: TABLE DATA; Schema: sysuser; Owner: revaananda
--

COPY sysuser.token (user_id, token, tstamp) FROM stdin;
12	a230d82c42c9ec0c8e41059a7ba45a1ae704d78864e48a811c3d20db70b18e6a	1739281899
\.


--
-- Data for Name: user; Type: TABLE DATA; Schema: sysuser; Owner: revaananda
--

COPY sysuser."user" (username, full_name, st, salt, saltedpassword, data, id, role) FROM stdin;
master	Master User	1	3w1WjEyeRFiYiQaB	3ee0c18f1887594557e3e2884e9d8a9c54cb5371e42769a3f4b12dda522ec5cd	{}	12	guest
admin1	admin1	1	RSf0HrwQV08kT04P	46d2c0ee31a2483810ddfe871da542472f208b54698dc4bdb7d6ab8dcf22ee4b	{}	13	guest
\.


--
-- Name: data_id_seq; Type: SEQUENCE SET; Schema: device; Owner: revaananda
--

SELECT pg_catalog.setval('device.data_id_seq', 357, true);


--
-- Name: device_id_sq; Type: SEQUENCE SET; Schema: public; Owner: revaananda
--

SELECT pg_catalog.setval('public.device_id_sq', 2, true);


--
-- Name: user_id_seq; Type: SEQUENCE SET; Schema: sysuser; Owner: revaananda
--

SELECT pg_catalog.setval('sysuser.user_id_seq', 13, true);


--
-- Name: data data_new_pkey1; Type: CONSTRAINT; Schema: device; Owner: revaananda
--

ALTER TABLE ONLY device.data
    ADD CONSTRAINT data_new_pkey1 PRIMARY KEY (id);


--
-- Name: unit unit_pkey; Type: CONSTRAINT; Schema: device; Owner: revaananda
--

ALTER TABLE ONLY device.unit
    ADD CONSTRAINT unit_pkey PRIMARY KEY (id);


--
-- Name: session session_pkey; Type: CONSTRAINT; Schema: sysuser; Owner: revaananda
--

ALTER TABLE ONLY sysuser.session
    ADD CONSTRAINT session_pkey PRIMARY KEY (session_id);


--
-- Name: session session_user_id_key; Type: CONSTRAINT; Schema: sysuser; Owner: revaananda
--

ALTER TABLE ONLY sysuser.session
    ADD CONSTRAINT session_user_id_key UNIQUE (user_id);


--
-- Name: token token_pkey; Type: CONSTRAINT; Schema: sysuser; Owner: revaananda
--

ALTER TABLE ONLY sysuser.token
    ADD CONSTRAINT token_pkey PRIMARY KEY (user_id, token);


--
-- Name: token unique_user_id; Type: CONSTRAINT; Schema: sysuser; Owner: revaananda
--

ALTER TABLE ONLY sysuser.token
    ADD CONSTRAINT unique_user_id UNIQUE (user_id);


--
-- Name: user user_pkey; Type: CONSTRAINT; Schema: sysuser; Owner: revaananda
--

ALTER TABLE ONLY sysuser."user"
    ADD CONSTRAINT user_pkey PRIMARY KEY (id);


--
-- Name: user user_unique_name; Type: CONSTRAINT; Schema: sysuser; Owner: revaananda
--

ALTER TABLE ONLY sysuser."user"
    ADD CONSTRAINT user_unique_name UNIQUE (username);


--
-- Name: data fk_unit; Type: FK CONSTRAINT; Schema: device; Owner: revaananda
--

ALTER TABLE ONLY device.data
    ADD CONSTRAINT fk_unit FOREIGN KEY (unit_id) REFERENCES device.unit(id) ON DELETE CASCADE;


--
-- Name: token fk_user_id; Type: FK CONSTRAINT; Schema: sysuser; Owner: revaananda
--

ALTER TABLE ONLY sysuser.token
    ADD CONSTRAINT fk_user_id FOREIGN KEY (user_id) REFERENCES sysuser."user"(id) ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--

