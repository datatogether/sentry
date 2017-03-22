package archive

const QSubprimerUndescribedContent = `
select
  urls.url, urls.created, urls.updated, urls.last_head, urls.last_get, urls.status, urls.content_type, urls.content_sniff, urls.content_length, 
  urls.title, urls.id, urls.headers_took, urls.download_took, urls.headers, urls.meta, urls.hash 
from urls, metadata 
where 
  urls.url ilike $1
  and content_sniff != 'text/html; charset=utf-8'
  and urls.last_get is not null
  -- confirm is not empty hash
  and urls.hash != '1220e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855'
  and not exists (select null from metadata where urls.hash = metadata.subject) 
limit $2 offset $3;`
