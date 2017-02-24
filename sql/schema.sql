DROP TABLE IF EXISTS urls, links, domains, alerts, supress_alerts;

CREATE TABLE domains (
	host 						text PRIMARY KEY NOT NULL,
	created 				integer NOT NULL,
	updated 				integer NOT NULL,
	stale_duration 	integer NOT NULL DEFAULT 43200000, -- defaults to 12 hours, column needs to be multiplied by 1000000 to become a poper duration
	crawl 					boolean default true,
	last_alert_sent bigint default 0
);

CREATE TABLE urls (
	url 						text PRIMARY KEY NOT NULL,
	created 				integer NOT NULL,
	updated 				integer NOT NULL,
	last_get 				integer NOT NULL default 0,
	host 						text,
	status 					integer default 0,
	content_type 		text default '',
	content_length 	bigint default 0,
	title  					text default '',
	id 							text default '',
	headers_took 		integer default 0,
	download_took 	integer default 0,
	headers 				json,
	meta 						json,
	hash 						text default ''
);

CREATE TABLE links (
	created 				integer NOT NULL,
	updated 				integer NOT NULL,
	src 						text references urls(url),
	dst 						text references urls(url),
	PRIMARY KEY 		(src, dst)
);

CREATE TABLE context (
	url 						text PRIMARY KEY NOT NULL references urls(url),
	created 				integer NOT NULL,
	updated 				integer NOT NULL,
	hash 						text NOT NULL,
	contributor_id 	text NOT NULL,
	context 				json
);

-- for domains table later?
-- cancelAfter = flag.Duration("cancelafter", 0, "automatically cancel the fetchbot after a given time")
-- cancelAtURL = flag.String("cancelat", "", "automatically cancel the fetchbot at a given URL")
-- stopAfter   = flag.Duration("stopafter", 0, "automatically stop the fetchbot after a given time")
-- stopAtURL   = flag.String("stopat", "", "automatically stop the fetchbot at a given URL")

-- CREATE TABLE captures (
-- 	id
-- );

-- CREATE TABLE alerts (
-- 	id 					UUID UNIQUE NOT NULL,
-- 	created 		integer NOT NULL,
-- 	updated 		integer NOT NULL,
-- 	dismissed 	boolean default false,
-- 	domain 			UUID references domains(id),
-- 	message 		text
-- );