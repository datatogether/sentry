
-- name: insert-domains
insert into domains values
	('www.epa.gov','2017-01-01 00:00:01','2017-01-01 00:00:01',43200000,true,null);

-- name: insert-urls
insert into urls values
	-- url,created,updated,last_get,last_head,host,status,content_type,content_length,title,id,headers_took,download_took,headers,meta,hash
	('http://www.epa.gov', '2017-01-01 00:00:01', '2017-01-01 00:00:01', '2017-01-01 00:00:01', null, 200, 'text/html; charset=utf-8', 'text/html;',-1, 'United States Environmental Protection Agency, US EPA', 'cee7bbd4-2bf9-4b83-b2c8-be6aeb70e771',0,0, '["X-Content-Type-Options","nosniff","Expires","Fri, 24 Feb 2017 21:53:45 GMT","Date","Fri, 24 Feb 2017 21:53:45 GMT","Etag","W/\"7f53-549471782bb42\"","X-Ua-Compatible","IE=Edge,chrome=1","X-Cached-By","Boost","Content-Type","text/html; charset=utf-8","Vary","Accept-Encoding","Accept-Ranges","bytes","Cache-Control","no-cache, no-store, must-revalidate, post-check=0, pre-check=0","Server","Apache","Connection","keep-alive","Strict-Transport-Security","max-age=31536000; preload;"]', null, '1220459219b10032cc86dcdbc0f83aea15a9d3e1119e7b5170beaee233008ea2c2de');

-- name: insert-links
-- insert into links values
-- ('2017-01-01 00:00:02','2017-01-01 00:00:02','http://www.epa.gov','http://www.epa.gov');

-- name: insert-snapshots
-- insert into snapshots values
-- 	();

-- name: insert-context
insert into context values
	-- url, contributor_id, created, updated, hash, meta
	('http://www.epa.gov','al','2017-01-01 00:00:04','2017-01-01 00:00:04','1220459219b10032cc86dcdbc0f83aea15a9d3e1119e7b5170beaee233008ea2c2de', '{ "title" : "EPA" }');

-- name: delete-domains
delete from domains;
-- name: delete-urls
delete from urls;
-- name: delete-links
delete from links;
-- name: delete-context
delete from context;
-- name: delete-snapshots
delete from snapshots;
