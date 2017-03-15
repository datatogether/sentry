
-- name: insert-primers
insert into primers values
  ('5b1031f4-38a8-40b3-be91-c324bf686a87','2017-01-01 00:00:01','2017-01-01 00:00:01', 'Environmental Protection Agency', 'The mission of the Environmental Protection Agency is to protect human health and the environment through the development and enforcement of regulations. The EPA is responsible for administering a number of laws that span various sectors, such as agriculture, transportation, utilities, construction, and oil and gas. In the budget for FY 2017, the agency lays out goals to better support communities and address climate change following the President’s Climate Action Plan. Additionally, the agency aims to improve community water infrastructure, chemical plant safety, and collaborative partnerships among federal, state, and tribal levels.',false);
-- name: delete-primers
delete from primers;

--name: insert-crawl_urls
insert into crawl_urls values
  ('www.epa.gov', '2017-01-01 00:00:01', '2017-01-01 00:00:01', '5b1031f4-38a8-40b3-be91-c324bf686a87',true,43200000,null,null);
--name: delete-crawl_urls
delete from crawl_urls;


-- name: insert-urls
insert into urls values
	-- url,created,updated,last_get,last_head,host,status,content_type,content_length,title,id,headers_took,download_took,headers,meta,hash
	('http://www.epa.gov', '2017-01-01 00:00:01', '2017-01-01 00:00:01', '2017-01-01 00:00:01', null, 200, 'text/html; charset=utf-8', 'text/html;',-1, 'United States Environmental Protection Agency, US EPA', 'cee7bbd4-2bf9-4b83-b2c8-be6aeb70e771',0,0, '["X-Content-Type-Options","nosniff","Expires","Fri, 24 Feb 2017 21:53:45 GMT","Date","Fri, 24 Feb 2017 21:53:45 GMT","Etag","W/\"7f53-549471782bb42\"","X-Ua-Compatible","IE=Edge,chrome=1","X-Cached-By","Boost","Content-Type","text/html; charset=utf-8","Vary","Accept-Encoding","Accept-Ranges","bytes","Cache-Control","no-cache, no-store, must-revalidate, post-check=0, pre-check=0","Server","Apache","Connection","keep-alive","Strict-Transport-Security","max-age=31536000; preload;"]', null, '1220459219b10032cc86dcdbc0f83aea15a9d3e1119e7b5170beaee233008ea2c2de');
  -- name: delete-urls
delete from urls;

-- name: insert-links
-- insert into links values
-- ('2017-01-01 00:00:02','2017-01-01 00:00:02','http://www.epa.gov','http://www.epa.gov');
-- name: delete-links
delete from links;
  
-- name: insert-snapshots
-- insert into snapshots values
-- 	();
-- name: delete-snapshots
delete from snapshots;

-- name: insert-context
insert into context values
  -- url, contributor_id, created, updated, hash, meta
  ('http://www.epa.gov','al','2017-01-01 00:00:04','2017-01-01 00:00:04','1220459219b10032cc86dcdbc0f83aea15a9d3e1119e7b5170beaee233008ea2c2de', '{ "title" : "EPA" }');
-- name: delete-context
delete from context;

