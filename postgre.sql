DROP DATABASE suggest;
CREATE DATABASE suggest;
\c suggest;

CREATE TABLE data (
  id          integer NOT NULL,
  pinyin      character varying(255) NOT NULL,
  abbr        character varying(255) NOT NULL,
  word        character varying(255) NOT NULL,
  sogou_id    integer NOT NULL,
  sogou_count integer DEFAULT 0 NOT NULL
);
CREATE SEQUENCE data_id_seq START WITH 1 INCREMENT BY 1 NO MINVALUE NO MAXVALUE CACHE 1;
ALTER SEQUENCE data_id_seq OWNED BY data.id;
ALTER TABLE ONLY data ADD CONSTRAINT data_pkey PRIMARY KEY (id);
ALTER TABLE ONLY data ALTER COLUMN id SET DEFAULT nextval('data_id_seq'::regclass);
CREATE INDEX index_data_on_pinyin ON data USING btree (pinyin);
CREATE INDEX index_data_on_abbr ON data USING btree (abbr);
CREATE INDEX index_data_on_word ON data USING btree (word);

CREATE TABLE categories (
  id                integer NOT NULL,
  sogou_category_id integer NOT NULL,
  name              character varying(255) NOT NULL
);
CREATE SEQUENCE categories_id_seq START WITH 1 INCREMENT BY 1 NO MINVALUE NO MAXVALUE CACHE 1;
ALTER SEQUENCE categories_id_seq OWNED BY categories.id;
ALTER TABLE ONLY categories ADD CONSTRAINT categories_pkey PRIMARY KEY (id);
ALTER TABLE ONLY categories ALTER COLUMN id SET DEFAULT nextval('categories_id_seq'::regclass);
CREATE UNIQUE INDEX index_categories_on_sogou_categories_id ON categories USING btree (sogou_category_id);

CREATE TABLE dicts (
  id             integer NOT NULL,
  sogou_id       integer NOT NULL,
  category_id    integer NOT NULL,
  name           character varying(255) NOT NULL,
  download_count integer NOT NULL,
  examples       text,
  updated_at     timestamp without time zone NOT NULL
);
CREATE SEQUENCE dicts_id_seq START WITH 1 INCREMENT BY 1 NO MINVALUE NO MAXVALUE CACHE 1;
ALTER SEQUENCE dicts_id_seq OWNED BY dicts.id;
ALTER TABLE ONLY dicts ADD CONSTRAINT dicts_pkey PRIMARY KEY (id);
ALTER TABLE ONLY dicts ALTER COLUMN id SET DEFAULT nextval('dicts_id_seq'::regclass);

CREATE FUNCTION SCORE(character varying, character varying)
RETURNS integer AS $$
BEGIN
	RETURN 10000 - ((position($2 in $1) - 1) * 100) - (length($1) - length($2)) * 10;
END;
$$ LANGUAGE plpgsql;
