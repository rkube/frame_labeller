
# SPARTA tagger
A small web-based image tagging tool.

Run as
```
$ go run .
```

## Workflow
1. Open `http://localhost:8080`
2. Use the form on the top right to register as any user you want. Thanks, communism.
3. Use the form on the top left to fetch some data.
4. Now a shot navigation slide bar pops up. Slide it use look at a frame
5. Create a label and a comment for the currently visible frame
6. Click the submit on the bottom right. Voila, your data is in a database now.


## Architecture

It's all HTMX. User state is managed on the backend using 2 structs.

An app_context that stores mapping from session ids to users and a mapping
from username to state
```go
// app_context is the local context. It is created in main and passed to all http handlers.
type app_context struct {
	session_to_user map[string]string     // Map from session ids to user ids
	all_user_state  map[string]user_state // Map from user id to user_state
}
```

The user-state is just the current shot number, frame, and session id.

```go
// Keeps track where each user is at a given time
type user_state struct {
	shotnr             int32
	frame              int32
	current_session_id string
}
```

This state is used to make sure we are writing labels correctly.

The following components are rendered based on the server-state associated with a user:

* sparta-shot-info  - Holds information (timing, number of frames) on loaded shot
* sparta-shot-nav - Nav buttons for in-shot navigation
* sparta-frame - Heatmap plot of the current frame




## Database format

Data is stored in table `test01`:
```
sqlite> .schema test01
CREATE TABLE test01 (
"username" Text,
"shotnr" Int,
"framenr" Int,
"label" Text,
"comment" Text);
```


Open the sqlite file and find all labels:

```
$ sqlite3
SQLite version 3.39.5 2022-10-14 20:58:05
Enter ".help" for usage hints.
Connected to a transient in-memory database.
Use ".open FILENAME" to reopen on a persistent database.
sqlite> .open test.sql
sqlite> .tables
test01
sqlite> SELECT * from test01
   ...> ;
rk|123|1|a label|a comment
rk|241013041|12|my label|foo222
```