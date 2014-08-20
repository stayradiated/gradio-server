gradio-server
=============

GrooveShark API wrapper and music manager

## API

### GET /song/:ID.mp3

**Request**

- `ID`: GrooveShark song id

**Response**

Returns the song with the specified id.

### GET /search/:QUERY

**Request**

- `QUERY`: Search query

**Response**

Returns JSON object with matching songs.
