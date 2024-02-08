# Collaborative Sh💩t Posting Pipeline (CSPP)

## Usage

As a user of the pipeline, you simply need to perform an authenticated upload of an image with a caption. From there it will be processed and sent to the channel configured by the `cspp` service.

The `image_path` is the path to the image on the local file system. The `caption` is the text that will be sent with the image to the channel.

### Curl example

To use this via `curl` you can do the following:

```shell
curl -X POST \
  -F "image=@/path/to/file" \
  -F "caption=String you want to with the picture"  \
  -H "X-API-KEY: $API_KEY" \
 <service URI>
```

### Other tools

The structure for the upload is as follows if you are using a different tool:

Send the API key as a header `X-API-KEY`.

The payload using HTTP POST on the `/api` end is:

```json
{
  "image_path": "/path/to/image.jpg",
  "caption": "This is a caption",
}
```

### Error handling

If you send an invaliad payload, the server may accept it, attempt to processes
it and then fail. If that happens, your requeset is kept and moved aside for
either manual debugging or manual processing. You may not receive an error if
that happens because the upload succeeded, just not the processing.
