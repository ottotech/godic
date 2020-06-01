--
-- PostgreSQL database dump
--

-- Dumped from database version 11.5 (Debian 11.5-1.pgdg90+1)
-- Dumped by pg_dump version 11.5 (Debian 11.5-1.pgdg90+1)

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

--
-- Name: counting_option; Type: TYPE; Schema: public; Owner: otto
--

CREATE TYPE public.counting_option AS ENUM (
    'unit',
    'decimal'
);


ALTER TYPE public.counting_option OWNER TO otto;

--
-- Name: period; Type: TYPE; Schema: public; Owner: otto
--

CREATE TYPE public.period AS ENUM (
    'morning',
    'afternoon',
    'night'
);


ALTER TYPE public.period OWNER TO otto;

SET default_tablespace = '';

SET default_with_oids = false;

--
-- Name: address; Type: TABLE; Schema: public; Owner: otto
--

CREATE TABLE public.address (
    id integer NOT NULL,
    identifier character varying(200) NOT NULL,
    street character varying(200) NOT NULL,
    number character varying(200),
    reference text,
    "default" boolean DEFAULT false NOT NULL,
    user_id integer NOT NULL
);


ALTER TABLE public.address OWNER TO otto;

--
-- Name: address_id_seq; Type: SEQUENCE; Schema: public; Owner: otto
--

CREATE SEQUENCE public.address_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.address_id_seq OWNER TO otto;

--
-- Name: address_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: otto
--

ALTER SEQUENCE public.address_id_seq OWNED BY public.address.id;


--
-- Name: administrator; Type: TABLE; Schema: public; Owner: otto
--

CREATE TABLE public.administrator (
    user_id integer NOT NULL
);


ALTER TABLE public.administrator OWNER TO otto;

--
-- Name: category; Type: TABLE; Schema: public; Owner: otto
--

CREATE TABLE public.category (
    id integer NOT NULL,
    name character varying(200) NOT NULL
);


ALTER TABLE public.category OWNER TO otto;

--
-- Name: category_id_seq; Type: SEQUENCE; Schema: public; Owner: otto
--

CREATE SEQUENCE public.category_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.category_id_seq OWNER TO otto;

--
-- Name: category_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: otto
--

ALTER SEQUENCE public.category_id_seq OWNED BY public.category.id;


--
-- Name: configuration; Type: TABLE; Schema: public; Owner: otto
--

CREATE TABLE public.configuration (
    delivery_capacity integer NOT NULL,
    delivery_window integer NOT NULL,
    min_purchase_amount double precision NOT NULL,
    iva double precision NOT NULL
);


ALTER TABLE public.configuration OWNER TO otto;

--
-- Name: invoice_detail; Type: TABLE; Schema: public; Owner: otto
--

CREATE TABLE public.invoice_detail (
    ruc character varying(200) NOT NULL,
    name character varying(200) NOT NULL,
    address character varying(200) NOT NULL,
    telephone character varying(200) NOT NULL,
    user_id integer NOT NULL,
    id integer NOT NULL
);


ALTER TABLE public.invoice_detail OWNER TO otto;

--
-- Name: invoice_detail_id_seq; Type: SEQUENCE; Schema: public; Owner: otto
--

CREATE SEQUENCE public.invoice_detail_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.invoice_detail_id_seq OWNER TO otto;

--
-- Name: invoice_detail_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: otto
--

ALTER SEQUENCE public.invoice_detail_id_seq OWNED BY public.invoice_detail.id;


--
-- Name: order; Type: TABLE; Schema: public; Owner: otto
--

CREATE TABLE public."order" (
    id integer NOT NULL,
    date_created timestamp without time zone NOT NULL,
    delivered boolean DEFAULT false NOT NULL,
    delivery_date timestamp without time zone NOT NULL,
    period public.period NOT NULL,
    iva double precision NOT NULL,
    user_id integer NOT NULL
);


ALTER TABLE public."order" OWNER TO otto;

--
-- Name: order_assignment; Type: TABLE; Schema: public; Owner: otto
--

CREATE TABLE public.order_assignment (
    id integer NOT NULL,
    date timestamp without time zone NOT NULL,
    from_time time without time zone,
    to_time time without time zone,
    order_id integer NOT NULL
);


ALTER TABLE public.order_assignment OWNER TO otto;

--
-- Name: order_assignment_id_seq; Type: SEQUENCE; Schema: public; Owner: otto
--

CREATE SEQUENCE public.order_assignment_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.order_assignment_id_seq OWNER TO otto;

--
-- Name: order_assignment_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: otto
--

ALTER SEQUENCE public.order_assignment_id_seq OWNED BY public.order_assignment.id;


--
-- Name: order_id_seq; Type: SEQUENCE; Schema: public; Owner: otto
--

CREATE SEQUENCE public.order_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.order_id_seq OWNER TO otto;

--
-- Name: order_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: otto
--

ALTER SEQUENCE public.order_id_seq OWNED BY public."order".id;


--
-- Name: order_line; Type: TABLE; Schema: public; Owner: otto
--

CREATE TABLE public.order_line (
    id integer NOT NULL,
    order_id integer NOT NULL,
    product_id integer NOT NULL,
    quantity_ordered double precision NOT NULL,
    quantity_delivered double precision NOT NULL,
    price double precision NOT NULL
);


ALTER TABLE public.order_line OWNER TO otto;

--
-- Name: order_line_id_seq; Type: SEQUENCE; Schema: public; Owner: otto
--

CREATE SEQUENCE public.order_line_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.order_line_id_seq OWNER TO otto;

--
-- Name: order_line_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: otto
--

ALTER SEQUENCE public.order_line_id_seq OWNED BY public.order_line.id;


--
-- Name: product; Type: TABLE; Schema: public; Owner: otto
--

CREATE TABLE public.product (
    id integer NOT NULL,
    product_code character varying(200) NOT NULL,
    name character varying(200) NOT NULL,
    description text NOT NULL,
    price double precision NOT NULL,
    thumbnail character varying(200) NOT NULL,
    active boolean DEFAULT true NOT NULL,
    has_iva boolean NOT NULL,
    brand_id integer NOT NULL,
    counting_option public.counting_option NOT NULL
);


ALTER TABLE public.product OWNER TO otto;

--
-- Name: product_brand; Type: TABLE; Schema: public; Owner: otto
--

CREATE TABLE public.product_brand (
    id integer NOT NULL,
    name character varying(200) NOT NULL
);


ALTER TABLE public.product_brand OWNER TO otto;

--
-- Name: product_brand_id_seq; Type: SEQUENCE; Schema: public; Owner: otto
--

CREATE SEQUENCE public.product_brand_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.product_brand_id_seq OWNER TO otto;

--
-- Name: product_brand_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: otto
--

ALTER SEQUENCE public.product_brand_id_seq OWNED BY public.product_brand.id;


--
-- Name: product_category; Type: TABLE; Schema: public; Owner: otto
--

CREATE TABLE public.product_category (
    product_id integer NOT NULL,
    category_id integer NOT NULL
);


ALTER TABLE public.product_category OWNER TO otto;

--
-- Name: product_id_seq; Type: SEQUENCE; Schema: public; Owner: otto
--

CREATE SEQUENCE public.product_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.product_id_seq OWNER TO otto;

--
-- Name: product_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: otto
--

ALTER SEQUENCE public.product_id_seq OWNED BY public.product.id;


--
-- Name: user; Type: TABLE; Schema: public; Owner: otto
--

CREATE TABLE public."user" (
    id integer NOT NULL,
    first_name character varying(80) NOT NULL,
    last_name character varying(80) NOT NULL,
    email character varying(200) NOT NULL,
    password character varying(200) NOT NULL,
    thumbnail character varying(200) NOT NULL,
    active boolean DEFAULT true NOT NULL,
    date_joined timestamp without time zone NOT NULL
);


ALTER TABLE public."user" OWNER TO otto;

--
-- Name: user_id_seq; Type: SEQUENCE; Schema: public; Owner: otto
--

CREATE SEQUENCE public.user_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.user_id_seq OWNER TO otto;

--
-- Name: user_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: otto
--

ALTER SEQUENCE public.user_id_seq OWNED BY public."user".id;


--
-- Name: address id; Type: DEFAULT; Schema: public; Owner: otto
--

ALTER TABLE ONLY public.address ALTER COLUMN id SET DEFAULT nextval('public.address_id_seq'::regclass);


--
-- Name: category id; Type: DEFAULT; Schema: public; Owner: otto
--

ALTER TABLE ONLY public.category ALTER COLUMN id SET DEFAULT nextval('public.category_id_seq'::regclass);


--
-- Name: invoice_detail id; Type: DEFAULT; Schema: public; Owner: otto
--

ALTER TABLE ONLY public.invoice_detail ALTER COLUMN id SET DEFAULT nextval('public.invoice_detail_id_seq'::regclass);


--
-- Name: order id; Type: DEFAULT; Schema: public; Owner: otto
--

ALTER TABLE ONLY public."order" ALTER COLUMN id SET DEFAULT nextval('public.order_id_seq'::regclass);


--
-- Name: order_assignment id; Type: DEFAULT; Schema: public; Owner: otto
--

ALTER TABLE ONLY public.order_assignment ALTER COLUMN id SET DEFAULT nextval('public.order_assignment_id_seq'::regclass);


--
-- Name: order_line id; Type: DEFAULT; Schema: public; Owner: otto
--

ALTER TABLE ONLY public.order_line ALTER COLUMN id SET DEFAULT nextval('public.order_line_id_seq'::regclass);


--
-- Name: product id; Type: DEFAULT; Schema: public; Owner: otto
--

ALTER TABLE ONLY public.product ALTER COLUMN id SET DEFAULT nextval('public.product_id_seq'::regclass);


--
-- Name: product_brand id; Type: DEFAULT; Schema: public; Owner: otto
--

ALTER TABLE ONLY public.product_brand ALTER COLUMN id SET DEFAULT nextval('public.product_brand_id_seq'::regclass);


--
-- Name: user id; Type: DEFAULT; Schema: public; Owner: otto
--

ALTER TABLE ONLY public."user" ALTER COLUMN id SET DEFAULT nextval('public.user_id_seq'::regclass);


--
-- Name: address address_pk; Type: CONSTRAINT; Schema: public; Owner: otto
--

ALTER TABLE ONLY public.address
    ADD CONSTRAINT address_pk PRIMARY KEY (id);


--
-- Name: category category_pk; Type: CONSTRAINT; Schema: public; Owner: otto
--

ALTER TABLE ONLY public.category
    ADD CONSTRAINT category_pk PRIMARY KEY (id);


--
-- Name: invoice_detail invoice_detail_pk; Type: CONSTRAINT; Schema: public; Owner: otto
--

ALTER TABLE ONLY public.invoice_detail
    ADD CONSTRAINT invoice_detail_pk PRIMARY KEY (id);


--
-- Name: order_assignment order_assignment_pk; Type: CONSTRAINT; Schema: public; Owner: otto
--

ALTER TABLE ONLY public.order_assignment
    ADD CONSTRAINT order_assignment_pk PRIMARY KEY (id);


--
-- Name: order_line order_line_pk; Type: CONSTRAINT; Schema: public; Owner: otto
--

ALTER TABLE ONLY public.order_line
    ADD CONSTRAINT order_line_pk PRIMARY KEY (id);


--
-- Name: order order_pk; Type: CONSTRAINT; Schema: public; Owner: otto
--

ALTER TABLE ONLY public."order"
    ADD CONSTRAINT order_pk PRIMARY KEY (id);


--
-- Name: product_brand product_brand_pk; Type: CONSTRAINT; Schema: public; Owner: otto
--

ALTER TABLE ONLY public.product_brand
    ADD CONSTRAINT product_brand_pk PRIMARY KEY (id);


--
-- Name: product_category product_category_product_id_category_id_key; Type: CONSTRAINT; Schema: public; Owner: otto
--

ALTER TABLE ONLY public.product_category
    ADD CONSTRAINT product_category_product_id_category_id_key UNIQUE (product_id, category_id);


--
-- Name: product product_pk; Type: CONSTRAINT; Schema: public; Owner: otto
--

ALTER TABLE ONLY public.product
    ADD CONSTRAINT product_pk PRIMARY KEY (id);


--
-- Name: user user_pk; Type: CONSTRAINT; Schema: public; Owner: otto
--

ALTER TABLE ONLY public."user"
    ADD CONSTRAINT user_pk PRIMARY KEY (id);


--
-- Name: address_id_uindex; Type: INDEX; Schema: public; Owner: otto
--

CREATE UNIQUE INDEX address_id_uindex ON public.address USING btree (id);


--
-- Name: address_identifier_uindex; Type: INDEX; Schema: public; Owner: otto
--

CREATE UNIQUE INDEX address_identifier_uindex ON public.address USING btree (identifier);


--
-- Name: administrator_user_id_uindex; Type: INDEX; Schema: public; Owner: otto
--

CREATE UNIQUE INDEX administrator_user_id_uindex ON public.administrator USING btree (user_id);


--
-- Name: category_id_uindex; Type: INDEX; Schema: public; Owner: otto
--

CREATE UNIQUE INDEX category_id_uindex ON public.category USING btree (id);


--
-- Name: category_name_uindex; Type: INDEX; Schema: public; Owner: otto
--

CREATE UNIQUE INDEX category_name_uindex ON public.category USING btree (name);


--
-- Name: order_assignment_id_uindex; Type: INDEX; Schema: public; Owner: otto
--

CREATE UNIQUE INDEX order_assignment_id_uindex ON public.order_assignment USING btree (id);


--
-- Name: order_id_uindex; Type: INDEX; Schema: public; Owner: otto
--

CREATE UNIQUE INDEX order_id_uindex ON public."order" USING btree (id);


--
-- Name: order_line_id_uindex; Type: INDEX; Schema: public; Owner: otto
--

CREATE UNIQUE INDEX order_line_id_uindex ON public.order_line USING btree (id);


--
-- Name: product_brand_id_uindex; Type: INDEX; Schema: public; Owner: otto
--

CREATE UNIQUE INDEX product_brand_id_uindex ON public.product_brand USING btree (id);


--
-- Name: product_brand_name_uindex; Type: INDEX; Schema: public; Owner: otto
--

CREATE UNIQUE INDEX product_brand_name_uindex ON public.product_brand USING btree (name);


--
-- Name: product_id_uindex; Type: INDEX; Schema: public; Owner: otto
--

CREATE UNIQUE INDEX product_id_uindex ON public.product USING btree (id);


--
-- Name: product_product_code_uindex; Type: INDEX; Schema: public; Owner: otto
--

CREATE UNIQUE INDEX product_product_code_uindex ON public.product USING btree (product_code);


--
-- Name: product_thumbnail_uindex; Type: INDEX; Schema: public; Owner: otto
--

CREATE UNIQUE INDEX product_thumbnail_uindex ON public.product USING btree (thumbnail);


--
-- Name: user_email_uindex; Type: INDEX; Schema: public; Owner: otto
--

CREATE UNIQUE INDEX user_email_uindex ON public."user" USING btree (email);


--
-- Name: user_id_uindex; Type: INDEX; Schema: public; Owner: otto
--

CREATE UNIQUE INDEX user_id_uindex ON public."user" USING btree (id);


--
-- Name: user_thumbnail_uindex; Type: INDEX; Schema: public; Owner: otto
--

CREATE UNIQUE INDEX user_thumbnail_uindex ON public."user" USING btree (thumbnail);


--
-- Name: address address_user_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: otto
--

ALTER TABLE ONLY public.address
    ADD CONSTRAINT address_user_id_fk FOREIGN KEY (user_id) REFERENCES public."user"(id) ON DELETE CASCADE;


--
-- Name: administrator administrator_user_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: otto
--

ALTER TABLE ONLY public.administrator
    ADD CONSTRAINT administrator_user_id_fk FOREIGN KEY (user_id) REFERENCES public."user"(id) ON DELETE CASCADE DEFERRABLE INITIALLY DEFERRED;


--
-- Name: invoice_detail invoice_detail_user_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: otto
--

ALTER TABLE ONLY public.invoice_detail
    ADD CONSTRAINT invoice_detail_user_id_fk FOREIGN KEY (user_id) REFERENCES public."user"(id) ON DELETE CASCADE;


--
-- Name: order_assignment order_assignment_order_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: otto
--

ALTER TABLE ONLY public.order_assignment
    ADD CONSTRAINT order_assignment_order_id_fk FOREIGN KEY (order_id) REFERENCES public."order"(id) ON DELETE CASCADE DEFERRABLE INITIALLY DEFERRED;


--
-- Name: order_line order_line_order_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: otto
--

ALTER TABLE ONLY public.order_line
    ADD CONSTRAINT order_line_order_id_fk FOREIGN KEY (order_id) REFERENCES public."order"(id) ON DELETE RESTRICT;


--
-- Name: order_line order_line_product_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: otto
--

ALTER TABLE ONLY public.order_line
    ADD CONSTRAINT order_line_product_id_fk FOREIGN KEY (product_id) REFERENCES public.product(id) ON DELETE RESTRICT;


--
-- Name: order order_user_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: otto
--

ALTER TABLE ONLY public."order"
    ADD CONSTRAINT order_user_id_fk FOREIGN KEY (user_id) REFERENCES public."user"(id) ON DELETE RESTRICT;


--
-- Name: product_category product_category_category_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: otto
--

ALTER TABLE ONLY public.product_category
    ADD CONSTRAINT product_category_category_id_fk FOREIGN KEY (category_id) REFERENCES public.category(id) ON DELETE RESTRICT;


--
-- Name: product_category product_category_product_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: otto
--

ALTER TABLE ONLY public.product_category
    ADD CONSTRAINT product_category_product_id_fk FOREIGN KEY (product_id) REFERENCES public.product(id) ON DELETE CASCADE;


--
-- Name: product product_product_brand_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: otto
--

ALTER TABLE ONLY public.product
    ADD CONSTRAINT product_product_brand_id_fk FOREIGN KEY (brand_id) REFERENCES public.product_brand(id) ON DELETE RESTRICT;


--
-- PostgreSQL database dump complete
--

