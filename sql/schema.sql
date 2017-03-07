-- name: drop-all
DROP TABLE IF EXISTS urls, links, domains, alerts, context, supress_alerts, snapshots;

-- name: create-domains
CREATE TABLE domains (
	host 						text PRIMARY KEY NOT NULL,
	created 				timestamp NOT NULL default (now() at time zone 'utc'),
	updated 				timestamp NOT NULL default (now() at time zone 'utc'),
	stale_duration 	integer NOT NULL DEFAULT 43200000, -- defaults to 12 hours, column needs to be multiplied by 1000000 to become a poper duration
	crawl 					boolean default true,
	last_alert_sent timestamp
);

-- name: create-urls
CREATE TABLE urls (
	url 						text PRIMARY KEY NOT NULL,
	created 				timestamp NOT NULL,
	updated 				timestamp NOT NULL,
	last_head 			timestamp,
	last_get 				timestamp,
	status 					integer NOT NULL default 0,
	content_type 		text NOT NULL default '',
	content_sniff 	text NOT NULL default '',
	content_length 	bigint NOT NULL default 0,
	title  					text NOT NULL default '',
	id 							text NOT NULL default '',
	headers_took 		integer NOT NULL default 0,
	download_took 	integer NOT NULL default 0,
	headers 				json,
	meta 						json,
	hash 						text NOT NULL default ''
);

-- name: create-links
CREATE TABLE links (
	created 				timestamp NOT NULL,
	updated 				timestamp NOT NULL,
	src 						text NOT NULL references urls(url) ON DELETE CASCADE,
	dst 						text NOT NULL references urls(url) ON DELETE CASCADE,
	PRIMARY KEY 		(src, dst)
);

-- name: create-context
CREATE TABLE context (
	url 						text NOT NULL references urls(url) ON DELETE CASCADE,
	contributor_id 	text NOT NULL,
	created 				timestamp NOT NULL,
	updated 				timestamp NOT NULL,
	hash 						text NOT NULL default '',
	meta 						json,
	UNIQUE 					(url, contributor_id)
);

-- name: create-snapshots
CREATE TABLE snapshots (
	url 						text NOT NULL references urls(url) ON DELETE CASCADE,
	created 				timestamp NOT NULL,
	status 					integer NOT NULL DEFAULT 0,
	duration 				integer NOT NULL DEFAULT 0,
	meta 						json,
	hash 						text NOT NULL DEFAULT ''
);

-- name: create-metablocks
-- CREATE TABLE metablocks (
-- 	time_stamp 			timestamp NOT NULL,
-- 	subject 				text NOT NULL,
-- 	meta 						text NOT NULL default '',
-- 	prev 						text NOT NULL default '',
-- 	key_id 					text NOT NULL default ''
-- );

-- for domains table later?
-- cancelAfter = flag.Duration("cancelafter", 0, "automatically cancel the fetchbot after a given time")
-- cancelAtURL = flag.String("cancelat", "", "automatically cancel the fetchbot at a given URL")
-- stopAfter   = flag.Duration("stopafter", 0, "automatically stop the fetchbot after a given time")
-- stopAtURL   = flag.String("stopat", "", "automatically stop the fetchbot at a given URL")

-- CREATE TABLE snapshots (
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