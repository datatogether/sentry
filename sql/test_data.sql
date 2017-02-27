
-- name: insert-domains
insert into domains values
	('www.epa.gov',0,0,43200000,true,0);

-- name: insert-urls
insert into urls values
	-- url,created,updated,last_get,host,status,content_type,content_length,title,id,headers_took,download_took,headers,meta,hash
	('http://www.epa.gov', 1487973224, 1487973225, 1487973225, 200, 'text/html; charset=utf-8',-1, 'United States Environmental Protection Agency, US EPA', 'cee7bbd4-2bf9-4b83-b2c8-be6aeb70e771',0,0, '["X-Content-Type-Options","nosniff","Expires","Fri, 24 Feb 2017 21:53:45 GMT","Date","Fri, 24 Feb 2017 21:53:45 GMT","Etag","W/\"7f53-549471782bb42\"","X-Ua-Compatible","IE=Edge,chrome=1","X-Cached-By","Boost","Content-Type","text/html; charset=utf-8","Vary","Accept-Encoding","Accept-Ranges","bytes","Cache-Control","no-cache, no-store, must-revalidate, post-check=0, pre-check=0","Server","Apache","Connection","keep-alive","Strict-Transport-Security","max-age=31536000; preload;"]', null, '1220459219b10032cc86dcdbc0f83aea15a9d3e1119e7b5170beaee233008ea2c2de');

-- name: insert-links
insert into links values
	(1487973225,1487973225,'http://www.epa.gov','http://www.epa.gov');

-- name: insert-captures
-- insert into captures values
-- 	();

-- name: insert-context
insert into context values
	('http://www.epa.gov','al',1487973225,1487973225,'1220459219b10032cc86dcdbc0f83aea15a9d3e1119e7b5170beaee233008ea2c2de', '{ "title" : "EPA" }');

-- name: delete-domains
delete from domains;
-- name: delete-urls
delete from urls;
-- name: delete-links
delete from links;
-- name: delete-context
delete from context;
-- name: delete-captures
delete from captures;
