# Slumbler API Server specification

---

User problem: 

Users have many links they see in a day. They'd like to be able to aggregate them, tag them, classify them, seach them and expire out links that are no longer valid or valuable.

This application "slumbler", will allow users to do just that. This is only the API server. Clients will be built separately.


User stories:
* As a user, I want to be able to add a link to my collection
* As a user, I want to be able to tag a link with one or more tags
* As a user, I want to be able to delete a tag associated with a link
* As a user, I want to be able to search for a link by tag
* As a user, I want to be able to expire a link manually
* As a user, I want links that a 404 to no longer be saved or stored.
* As a user, I want to be able to see the most popular links
* As a user, I want to be able to see the most recent links
* As a user, I want to be able to see the most popular tags
* As a user, I want to be able to see the most recent tags


Some technical requirements:
* The application should be a RESTful API written in Go
* The application should have a /healthz endpoint that returns 200 OK
* The application should have a /metrics endpoint that returns prometheus metrics
* The /metrics endpoint will include counters for requests, response status codes, link views, and link creation attempts.
* The application should have a /links endpoint that returns all links
* The application should have a /link/{id} endpoint that returns a single link
* The application should have a /link/{id}/tags endpoint that returns all tags for a link
* The application should have a /tags endpoint that returns all tags
* The application should have a /popular endpoint that returns the most popular links
* The application should have a /recent endpoint that returns the most recent links
* The application should have a /popular-tags endpoint that returns the most popular tags
* The application should have a /recent-tags endpoint that returns the most recent tags
* All endpoints should return JSON
* API results should be paginated with a default of 30 results per page
* The parameter for pagination should be `page`
* The parameter for pagination should be `per_page`
* The maximum number of results per page should be 100
* For API convention, the URIs should be plural (e.g. /links) when retreiving a collection and singular (e.g. /link/{id}) when retreiving a single item
* Database definitions should be in a separate file that can be run via migrations of some kind. 
* Respones should be in JSON, return total result count, current page and total number of pages
* The API should be versioned in the URL (e.g. /v1/links) and should be able to support multiple versions at the same time
* THe user exposed API may be presented with simply /links, /tags, /tag/{id}, /link/{id} and /search endpoints, as we can proxy that to the versioned API
* The API should be in OpenAPI 3.0 format with documentation in the code and swagger generated from the code
* The API docs generated should have examples using `curl` for each endpoint
* API Docs should be at the /apidoc endpoint


Database
---
* The database engine should be SQLite for now. In the future we'll likely use MySQL.
* Migrations should be in a migrations directory that can be run with a command or happen automatically when the application starts
* Rows added to the database should have a created_at and updated_at timestamp
The owner field is currently a string, but in future versions may reference a UUID or user ID. It should not be assumed to be an email address or human-readable value.

Tags
---
* Tags are case insensitive
* Tags should be saved and displayed as case-sensitive, but searched for case-insensitive
* In the event a new tag is created with different capitalization, do not create the new tag and inform the user the tag already exists
* Tags are global
* There can be up to 64 tags per link
* IF a user attempts to add more than 64 tags to a link, a 400 Bad Request should be returned
* If a user seraches for a tag that is cased differently, it should be returned how it is stored. (a user searching for a tag 'devops' would see 'DevOps' if that's how the tag was originally created)
* When searching for tags, if a user is searching for multiple tags (.e.g tag:go,tag:devops) the comma should be treated as an AND operator. A space would be an OR operator. 
* Syntax of tag:go,devops is an AND operator. Syntax of tag:go devops is an OR operator.
* When a new tag is submitted with a casing conflict, the request will be accepted, and the canonical casing will be returned. No duplicates will be created.
* A tag is removed the system entirely if it is not associated with any links

Search
---
* Search should be case insensitive
* Search should be a substring search
* Search should be able to search for multiple tags
* Search should be able to fuzzy match titles and tags
* Search should be able to search for a link by title
* Search should be able to search for a link by URL
* Search should be able to search for a link by title and URL
* Search can be FTS5 for SQLite
* If user does not want to provide a field to search, it should search title (or opengraph title), then URL, then tags. 
* To search a specific field, use the field name as a prefix to the search term. For example, `tag:mytag` or `title:mytitle` or `url:myurl`, but the values provided after the field deliter should be case insensitive and still fuzzy matched.
* Generically, search should follow the pattern outlined in the "Tags" section for other fields such as title:, url:, etc
* Spaces around commas in search syntax (e.g. tag:go, devops) will be treated as an OR unless explicitly parsed. Clients are encouraged to avoid mixed delimiters.

## Link-Tag Relationship Endpoints

The relationship between links and tags is many-to-many. To manage this relationship, the API supports the following **link-centric endpoints** and corresponding HTTP verbs. These endpoints should be used to add, remove, retrieve, and update tags associated with a given link.

### 🔗 Link-Centric Tag Management

| Endpoint                       | HTTP Verb | Description                                                       |
|-------------------------------|-----------|-------------------------------------------------------------------|
| `/link/{link_id}/tags`        | GET       | List all tags associated with the specified link.                 |
| `/link/{link_id}/tags`        | POST      | Add one or more tags to the specified link.                       |
| `/link/{link_id}/tag/{tag_id}`| GET       | Retrieve a single tag associated with the link.                   |
| `/link/{link_id}/tag/{tag_id}`| DELETE    | Remove the specified tag from the link.                           |
| `/link/{link_id}/tag/{tag_id}`| PUT       | Update the tag on the link (e.g. reassign a tag reference).       |

> **Note:** Adding tags via `POST /link/{link_id}/tags` can accept a list of tags in the request body (e.g. `["Go", "DevOps"]`). Tags will be deduplicated and case-insensitive during creation.  
> **Note:** A maximum of 64 tags per link is enforced.

---

### 🏷️ Tag-Centric Link Retrieval (Read-Only)

Tag-centric endpoints are available for retrieving links associated with a specific tag but should not be used for modifying the relationship.

| Endpoint                         | HTTP Verb | Description                                                      |
|----------------------------------|-----------|------------------------------------------------------------------|
| `/tag/{tag_id}/links`           | GET       | List all links associated with the specified tag.                |
| `/tag/{tag_id}/link/{link_id}`  | GET       | Retrieve a single link associated with the specified tag.        |


Links
---
* When a link is is stored in the database, it should also record the timestamp
* When a link is stored in the database, it should also record the user who added the link if that is known
* Prior to saving a link, it should be checked to ensure it can be reached with a 200 OK status code (including redirects)
* If a link is not reachable, it should not be saved
* Expiring a link doesn't delete it from the database, but it should not be returned in the /links endpoint
* Expiring a link should be reversible
* The title of the web page should be stored with the link if it is available (meaning there will be some HTML scraping)
* The link should store it's opengraph data if it's available. 
* Retrieving of opengraph data should be on the /link/{id} endpoint in a nested data structure.
* To get the title, prefer opengraph, then fall back to twitter card tags, then to scraping for the html title tag
* Opengraph or other title information should be stored in the database. It could be updated/refreshed using a PUT to the /link/{id} endpoint
* If a link is attempted to be added that already exists, the existing link should be returned and say when it was added.
* Links are unique per user, but in the initial implementation they are unique in the whole system
* Uniqueness of links should be link + user. Other metadata is not considered in the uniqueness of a link.
* To unexpire a link (which will likely be rare), you can use an HTTP PATCH request to /link/{id}
* For popularity of links, increment a counter in the database for each "get" of a link. This will be a separate table in the database and should be timstamped so a user could search for what's popular in the last week, month, etc.
* For 404 checking, redirects should be followed including 301 and 302s. If a 404 is found, the link should be expired or not created.
* Redirects should reset the timeout for 404 checking, but only for 4 redirects. After 4 redirects or 4 seconds it should stop trying and not add the link.
* For 404 checking, timeout after 4 seconds.
* If a timeout happens, warn the user in the response, and don't return a success. 
* If a link is already in the database, but the title is different, update the title in the database. If it 404s, expire the link and inform the user.
* When a link is stored the first time, it should store title and opengraph data if available.
* A PUT on /link/{id} should refresh metadata for the link, including the title and opengraph data
* A successful PUT /link/{id} returns the refreshed metadata in the response. If the link is expired due to a 404, the response includes an error and the updated metadata is not returned.
* A GET on /link/{id} should increase the view count, even if it's the same user or IP address. 
* PUT on /link/{id} should not increase the view count
* PUT on /link/{id} should not change the created_at timestamp
* A successful PUT /link/{id} returns the full link object, including tags, created_at, view_count, and opengraph, even if no fields changed.
* Metadata refresh should be triggered when a new link is being added to the application or when a PUT request is explicitly called. There is no other time it needs to happen (e.g. a GET request should not trigger a metadata refresh)
* If a PUT on /link/{id} now returns a 404, the link should be expired and the user should be informed. The title or Opengraph data should not be updated or changed.
* The only allowed PATCH operation in this version is { "expired": false } to unexpire a link.
* If a user attempts to add a link that already exists, but has a different set of tags, those new tags should be added to the existing link.
* If the service can't get to a link due to authentication, it should not be saved and the user should be informed.
* A link may have zero tags. All tags are optional.
* If a user attempts to add a link that already exists but is expired, the API should return the existing link with a flag indicating it is expired. It is up to the client to explicitly unexpire the link via PATCH.




Scope specifics:
* Authentication is out of scope for the current iteration of the project
* There is no concept of users (in terms of IAM) in the current iteration of the project
* A job to periodically check for expired links is out of scope for the current iteration of the project
* For initial scope, since there are no users, there is no concept of sharing links
* In order to account for having users in the future, links should have an owner field that is NOT NULL. For global links (like when there is no users yet), the owner can be "system". 
* The reserved owner string "system" is used when user information is unavailable. This value is case-sensitive and should not be used by clients once user accounts are introduced.
* In this version, there are no rate limits or authentication. Future versions may introduce scoped access or limits based on token or IP.

Libaries and tools from Go:
* No prefernce for swagger/OpenAPI libraries. Use something modern that is still maintained.
* For migrations, use golang-migrate/migrate
* To do opengraph or web scraping, there is no preference, provided it is still being maintained
* For metrics, any prometheus library is fine if it's still maintained


# Sample Respones

### OpenGraph Structure

The substructure for the OpenGraph data for a link could modelled in a way that follows

```json
{
  "url": "https://example.com",
  "title": "Example Site",
  "opengraph": {
    "title": "OG Title",
    "description": "OG Desc",
    "image": "https://example.com/image.jpg",
    "site_name": "Example",
    "type": "website"
  }
}
```

### Too many Tags

```json
{
  "type": "https://example.com/probs/too-many-tags",
  "title": "Too many tags",
  "status": 400,
  "detail": "Links can have at most 64 tags"
}
```

### Search Results

```json
{
  "page": 1,
  "per_page": 30,
  "total_results": 92,
  "total_pages": 4,
  "results": [
    {
      "id": "abc123",
      "url": "https://example.com",
      "title": "Example",
      "tags": ["Go", "DevOps"],
      "created_at": "2025-03-24T12:00:00Z",
      "updated_at": "2025-03-24T14:02:00Z",
      "expired": false,
      "view_count": 14,
      "opengraph": { ... }
    }
  ]
}
```





Future thoughts
---
* Click count for popularity of links
* random endpoint
* comments (after users)
* A browser exention to be a client as well
* A mobile app to be a client as well
* Import/export for links per person
* Analytics on links

