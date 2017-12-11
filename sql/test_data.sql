
-- name: insert-primers
insert into primers 
  (id,created,updated,short_title,title,description,parent_id,deleted)
values
  ('5b1031f4-38a8-40b3-be91-c324bf686a87','2017-01-01 00:00:01','2017-01-01 00:00:01', 'Localhost test', 'Its a web page with cute chinchilas :)', 'Lorem ipsum, just kidding.', '', false);
-- name: delete-primers
delete from primers;

--name: insert-sources
insert into sources
  (id,created,updated,title,description,url,primer_id,crawl,stale_duration,last_alert_sent,stats,meta)
values
  ('326fcfa0-d3e6-4b2d-8f95-e77220e16109', '2017-01-01 00:00:01', '2017-01-01 00:00:01', '127.0.0.1', 'entire localhost site', '127.0.0.1:8002', '5b1031f4-38a8-40b3-be91-c324bf686a87',true,43200000,null,null,null);
--name: delete-sources
delete from sources;

-- name: insert-urls

-- name: delete-urls
delete from urls;

