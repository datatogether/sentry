package archive

const qCollectionInsert = `
insert into collections 
  (id, created, updated, creator, title, schema, contents ) 
values ($1, $2, $3, $4, $5, $6, $7);`

const qCollectionUpdate = `
update collections 
set created=$2, updated=$3, creator=$4, title=$5, schema=$6, contents=$7 
where id = $1;`

const qCollectionById = `
select 
  id, created, updated, creator, title, schema, contents 
from collections 
where id = $1;`

const qCollectionDelete = `
delete from collections 
where id = $1;`

const qCollections = `
select
  id, created, updated, creator, title, schema, contents
from collections 
order by created desc 
limit $1 offset $2;`

const qMetadataLatest = `
select
  hash, time_stamp, key_id, subject, prev, meta 
from metadata 
where 
  key_id = $1 and 
  subject = $2 
order by time_stamp desc;`

const qMetadataForSubject = `
select
  hash, time_stamp, key_id, subject, prev, meta  
from metadata
where 
  subject = $1 and 
  deleted = false and 
  meta is not null;`

const qMetadataCountForKey = `
select
  count(1)
from metadata
where
  -- confirm is not empty hash
  hash != '1220e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855' and
  key_id = $1;`

const qMetadataLatestForKey = `
select distinct on (subject)
  hash, time_stamp, key_id, subject, prev, meta
from metadata
where
  key_id = $1 and
  deleted = false
order by subject, time_stamp desc
limit $2 offset $3;`

const qMetadataInsert = `
insert into metadata
  (hash, time_stamp, key_id, subject, prev, meta, deleted)
values 
  ($1, $2, $3, $4, $5, $6, false);`

const qPrimerById = `
select
  id, created, updated, short_title, title, description, 
  parent_id, stats, meta
from primers 
where 
  deleted = false and
  id = $1;`

const qPrimerInsert = `
insert into primers
  (id, created, updated, short_title, title, description, parent_id, stats, meta)
values
  ($1, $2, $3, $4, $5, $6, $7, $8, $9);`

const qPrimerUpdate = `
update primers set
  created = $2, updated = $3, short_title = $4, title = $5, description = $6,
  parent_id = $7, stats = $8, meta = $9
where
  deleted = false and
  id = $1;`

const qPrimerDelete = `
update primers set
  deleted = true
where id = $1;`

const qPrimerSubPrimers = `
select
  id, created, updated, short_title, title, description, 
  parent_id, stats, meta
from primers
where 
  deleted = false and
  parent_id = $1;`

const qPrimerSources = `
select
  id, created, updated, title, description, url, primer_id, crawl, stale_duration,
  last_alert_sent, meta, stats
from sources
where
  deleted = false and
  primer_id = $1;`

const qPrimersList = `
select
  id, created, updated, short_title, title, description,
  parent_id, stats, meta
from primers
where
  deleted = false
order by created desc
limit $1 offset $2;`

const qSourcesList = `
select
  id, created, updated, title, description, url, primer_id, crawl, stale_duration,
  last_alert_sent, meta, stats
from sources
where 
  deleted = false
order by created desc
limit $1 offset $2;`

const qSourceById = `
select
  id, created, updated, title, description, url, primer_id, crawl, stale_duration,
  last_alert_sent, meta, stats
from sources 
where
  deleted = false and
  id = $1;`

const qSourceByUrl = `
select
  id, created, updated, title, description, url, primer_id, crawl, stale_duration,
  last_alert_sent, meta, stats
from sources 
where
  deleted = false and
  url = $1;`

const qSourceInsert = `
insert into sources 
  (id, created, updated, title, description, url, primer_id, crawl, stale_duration,
   last_alert_sent, meta, stats) 
values 
  ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12);`

const qSourceUpdate = `
update sources 
set 
  created = $2, updated = $3, title = $4, description = $5, url = $6, primer_id = $7, 
  crawl = $8, stale_duration = $9, last_alert_sent = $10, meta = $11, stats = $12
where
  deleted = false and
  id = $1;`

const qSourceDelete = `
update sources
set
  deleted = true
where 
  url = $1;`

const qSourceUrlCount = `
select count(1) 
from urls 
where 
  url ilike $1;`

const qSourcesCrawling = `
select
  id, created, updated, title, description, url, parent_id, crawl, stale_duration,
  last_alert_sent, meta, stats
from sources
where
  deleted = false and
  crawl = true 
limit $1 offset $2;`

const qSourceCrawlingUrls = `
select
  id, created, updated, title, description, url, primer_id, crawl, stale_duration,
  last_alert_sent, meta, stats
from sources
where
  deleted = false and
  crawl = true;`

const qSourceContentUrlCount = `
select count(1) 
from urls 
where
  hash != '' and
  hash != '1220e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855' and
  content_sniff != 'text/html; charset=utf-8' and
  url ilike $1;`

const qSourceContentWithMetadataCount = `
select count(1)
from urls 
where 
  url ilike $1 and 
  content_sniff != 'text/html; charset=utf-8' and
  -- confirm is not empty hash
  hash != '1220e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855' and
  hash != '' and
  exists (select null from metadata where urls.hash = metadata.subject);`

const qSourceUndescribedContentUrls = `
select
  url, created, updated, last_head, last_get, status, content_type, content_sniff,
  content_length, file_name, title, id, headers_took, download_took, headers, meta, hash
from urls 
where 
  url ilike $1
  and content_sniff != 'text/html; charset=utf-8'
  and last_get is not null
  -- confirm is not empty hash
  and hash != '1220e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855'
  and not exists (select null from metadata where urls.hash = metadata.subject) 
limit $2 offset $3;`

const qSourceDescribedContentUrls = `
select
  url, created, updated, last_head, last_get, status, content_type, content_sniff,
  content_length, file_name, title, id, headers_took, download_took, headers, meta, hash
from urls 
where 
  url ilike $1
  and content_sniff != 'text/html; charset=utf-8'
  and last_get is not null
  -- confirm is not empty hash
  and hash != '1220e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855'
  and exists (select null from metadata where urls.hash = metadata.subject) 
limit $2 offset $3;`

const qSnapshotsByUrl = `
select
  url, created, status, duration, meta, hash
from snapshots 
where 
  url = $1;`

const qSnapshotInsert = `
insert into snapshots 
  (url, created, status, duration, meta, hash)
values 
  ($1, $2, $3, $4, $5, $6);`

const qUrlsSearch = `
select
  url, created, updated, last_head, last_get, status, content_type, content_sniff,
  content_length, file_name, title, id, headers_took, download_took, headers, meta, hash
from urls 
where 
  url ilike $1 
limit $2 offset $3;`

const qUrlsList = `
select
  url, created, updated, last_head, last_get, status, content_type, content_sniff,
  content_length, file_name, title, id, headers_took, download_took, headers, meta, hash
from urls 
order by created desc 
limit $1 offset $2;`

const qContentUrlsList = `
select
  url, created, updated, last_head, last_get, status, content_type, content_sniff,
  content_length, file_name, title, id, headers_took, download_took, headers, meta, hash
from urls 
where
  last_get is not null and
  content_sniff != 'text/html; charset=utf-8' and
  content_sniff != '' and
  hash != ''
order by created desc
limit $1 offset $2;`

const qUrlsFetched = `
select
  url, created, updated, last_head, last_get, status, content_type, content_sniff,
  content_length, file_name, title, id, headers_took, download_took, headers, meta, hash
from urls 
where
  last_get is not null 
order by created desc
limit $1 offset $2;`

const qUrlsUnfetched = `
select
  url, created, updated, last_head, last_get, status, content_type, content_sniff,
  content_length, file_name, title, id, headers_took, download_took, headers, meta, hash
from urls
where 
  last_get is null 
order by created desc 
limit $1 offset $2;`

const qUrlsForHash = `
select
  url, created, updated, last_head, last_get, status, content_type, content_sniff,
  content_length, file_name, title, id, headers_took, download_took, headers, meta, hash
from urls
where 
  hash = $1;`

const qUrlByUrlString = `
select
  url, created, updated, last_head, last_get, status, content_type, content_sniff,
  content_length, file_name, title, id, headers_took, download_took, headers, meta, hash
from urls 
where
  url = $1;`

const qUrlById = `
select
  url, created, updated, last_head, last_get, status, content_type, content_sniff,
  content_length, file_name, title, id, headers_took, download_took, headers, meta, hash
from urls 
where
  id = $1;`

const qUrlByHash = `
select
  url, created, updated, last_head, last_get, status, content_type, content_sniff,
  content_length, file_name, title, id, headers_took, download_took, headers, meta, hash
from urls 
where
  hash = $1;`

const qUrlInsert = `
insert into urls
  (url, created, updated, last_head, last_get, status, content_type, content_sniff,
  content_length, file_name, title, id, headers_took, download_took, headers, meta, hash)
values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17);`

const qUrlUpdate = `
update urls 
set 
  created=$2, updated=$3, last_head=$4, last_get=$5, status=$6, content_type=$7, content_sniff=$8,
  content_length=$9, file_name=$10, title=$11, id=$12, headers_took=$13, download_took=$14, headers=$15, meta=$16, hash=$17 
where 
  url = $1;`

const qUrlDelete = `
delete from urls 
where
  url = $1;`

const qUrlInboundLinkUrlStrings = `
select src 
from links 
where
  dst = $1;`

const qUrlOutboundLinkUrlStrings = `
select dst 
from links 
where
  src = $1;`

const qUrlDstLinks = `
select 
  urls.url, urls.created, urls.updated, last_head, last_get, status, content_type, content_sniff, 
  content_length, title, id, headers_took, download_took, headers, meta, hash 
from urls, links
where 
  links.src = $1 and 
  links.dst = urls.url;`

const qUrlSrcLinks = `
select
  urls.url, urls.created, urls.updated, last_head, last_get, status, content_type, content_sniff, 
  content_length, title, id, headers_took, download_took, headers, meta, hash 
from urls, links 
where 
  links.dst = $1 and 
  links.src = urls.url;`

const qUncrawlableInsert = `
insert into uncrawlables 
  ( url, created,updated,creator_key_id,
    name,email,event_name,agency_name,
    agency_id,subagency_id,org_id,suborg_id,subprimer_id,
    ftp,database,interactive,many_files,
    comments) 
values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18);`

const qUncrawlableUpdate = `
update uncrawlables 
set
  created = $2, updated = $3, creator_key_id = $4,
  name = $5, email = $6, event_name = $7, agency_name = $8,
  agency_id = $9, subagency_id = $10, org_id = $11, suborg_id = $12, subprimer_id = $13,
  ftp = $14, database = $15,interactive = $16, many_files = $17,
  comments = $18
where url = $1;`

const qUncrawlableByUrl = `
select 
  url,created,updated,creator_key_id,
  name,email,event_name,agency_name,
  agency_id,subagency_id,org_id,suborg_id,subprimer_id,
  ftp,database,interactive,many_files,
  comments
from uncrawlables 
where url = $1;`

const qUncrawlableDelete = `
delete from uncrawlables 
where url = $1;`

const qUncrawlables = `
select
  url,created,updated,creator_key_id,
  name,email,event_name,agency_name,
  agency_id,subagency_id,org_id,suborg_id,subprimer_id,
  ftp,database,interactive,many_files,
  comments
from uncrawlables 
order by created desc 
limit $1 offset $2;`
