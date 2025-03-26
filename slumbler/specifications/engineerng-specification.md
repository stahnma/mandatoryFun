# Slumbler API Specification

**Slumbler** is a RESTful API server designed to help users collect, tag, classify, and search for links — while intelligently handling link expiration and metadata extraction. This is the API backend only. Clients will be built separately.

---

## 🧩 User Problem

Users encounter many links each day and need a way to:

- Save them for later
- Classify with tags
- Search across titles, URLs, and tags
- See trending links
- Automatically expire links that no longer resolve (e.g., 404)

---

## 📋 User Stories

- ✅ Add a link to my collection
- ✅ Tag a link with one or more tags
- ✅ Delete a tag associated with a link
- ✅ Search for a link by tag
- ✅ Manually expire a link
- ✅ Prevent 404 links from being saved
- ✅ View the most popular links
- ✅ View the most recent links
- ✅ View the most popular tags
- ✅ View the most recent tags

---

## 📐 Technical Requirements

- Written in Go
- RESTful API, returns JSON
- OpenAPI 3.0 definition with Swagger support
- All endpoints versioned via URI prefix (e.g., `/v1/`)
- Pagination defaults to 30 results per page, max 100
- Paginated endpoints accept `page` and `per_page` query parameters
- Prometheus metrics at `/metrics`, including:
  - Request count
  - Response status codes
  - Link view counters
  - Link creation attempts
- Health check at `/healthz` returns 200 OK
- Swagger UI available at `/apidoc`

---

## 🗂️ Routes

| Method | Endpoint                           | Description |
|--------|------------------------------------|-------------|
| GET    | /healthz                           | Health check |
| GET    | /metrics                           | Prometheus metrics |
| GET    | /links                             | List all active links |
| POST   | /links                             | Add a new link |
| GET    | /link/{id}                         | Get link details (increments view count) |
| PUT    | /link/{id}                         | Refresh metadata (does not increment view count) |
| PATCH  | /link/{id}                         | Only `{ "expired": false }` supported to unexpire |
| GET    | /link/{id}/tags                    | List tags for a link |
| POST   | /link/{id}/tags                    | Add one or more tags |
| GET    | /link/{id}/tag/{tag_id}            | Get a specific tag |
| DELETE | /link/{id}/tag/{tag_id}            | Remove a tag from a link |
| PUT    | /link/{id}/tag/{tag_id}            | Update/reassign a tag |
| GET    | /tags                              | List all tags |
| GET    | /tag/{tag_id}/links                | List all links for a tag |
| GET    | /tag/{tag_id}/link/{link_id}       | Get specific link for tag |
| GET    | /popular                           | Most popular links (by view count) |
| GET    | /recent                            | Most recently added links |
| GET    | /popular-tags                      | Most popular tags |
| GET    | /recent-tags                       | Most recently used tags |
| GET    | /search                            | Search links by tag/title/url/etc. |

---

## 🧠 Link Storage Logic

- All links must resolve with HTTP 200 OK (after up to 4 redirects, 4s timeout)
- Failing to fetch due to auth or 404 results in link being rejected or expired
- On adding duplicate links:
  - If not expired: return existing record
  - If expired: return record with `"expired": true`
  - Any new tags are merged with the existing link
- Links are globally unique in current version
  - In future, uniqueness will be `url + user`
- Owner field is required (currently string, e.g., `"system"`)

---

## 🏷️ Tags

- Global tag namespace
- Case-insensitive matching; case-sensitive storage
- Max 64 tags per link
- Tags are deleted system-wide if unused
- Tag casing conflicts are accepted, but stored casing is returned
- Search syntax:
  - `tag:go,devops` → AND
  - `tag:go devops` → OR
  - Mixed delimiters (`tag:go, devops`) treated as OR

---

## 🔍 Search

- Case-insensitive, fuzzy, and substring matches
- Fields: title, URL, tag
- Can combine filters via `tag:`, `url:`, `title:`
- No field prefix: default search order is title → URL → tags
- Uses SQLite FTS5

---

## 🔁 Link Metadata

- Title extraction:
  1. OpenGraph
  2. Twitter Card
  3. `<title>` tag
- OpenGraph data is stored and returned as a sub-object
- Metadata only refreshed on:
  - Initial creation
  - PUT /link/{id}
- PUT returns entire updated link record
- PUT does not update `created_at`, nor increment view count
- GET /link/{id} increments view count (even for same IP/user)

---

## 🔄 Expiration Handling

- Link expiration:
  - Manual via PATCH
  - Auto-expire on 404 during metadata refresh or re-fetch
- Expired links are retained in DB but hidden from most list endpoints
- Only PATCH payload allowed: `{ "expired": false }`

---

## 🛠️ Libraries & Tools

- **Migrations:** [`golang-migrate/migrate`](https://github.com/golang-migrate/migrate)
- **OpenAPI:** No preference, must be maintained
- **Metrics:** Prometheus (any modern library)
- **Web scraping:** Any maintained HTML/OG library
- **Database:** SQLite (with migration path to MySQL)

---

## 🔐 Scope Specifics

- No authentication or user system (yet)
- `owner` field is required (`"system"` for now)
- No concept of sharing or permissions
- No scheduled revalidation jobs yet
- Future versions may:
  - Introduce user IDs
  - Add auth/rate limiting
  - Support scoped visibility or collaboration

---

## 📦 Example Responses

### ✅ Link with OpenGraph

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

### ❌ Too Many Tags

```json
{
  "type": "https://example.com/probs/too-many-tags",
  "title": "Too many tags",
  "status": 400,
  "detail": "Links can have at most 64 tags"
}
```

### 🔍 Search Results

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

---

## 🚧 Future Considerations

- Link click tracking
- Random link endpoint
- Comment support
- Browser extension
- Mobile client
- Import/export for links
- Link analytics
