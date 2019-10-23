
- State of transactions:
  - account
  - transaction itself

- State of authentication connection:
  - refresh token, token, clientID, clientSecret

- Historical state -- important for refreshing from the past.

Preferred approach state-wise:
 - Persisted record of all information stored in JSON.
 - Updated (merged) periodically.
 - Periodically merged into main book of records.

--------------------------------------------
Monzo downloader tool:
 - Keep a single JSON file that stores all auth configuration
 - Also keep downloaded state.
 - Automatically downloads everything since min(90 days ago, last update date).
--------------------------------------------
... but how is this done exactly?
 Where is the business logic?
  - For the previous example, there was a conversion script into ledger format.
  - Do something similar with ops?
  - Formatter does JSON-path extraction?

----------------------------------------------

goledger download

- monzo is built-in-by-code
- all in one directory
- auth URLs are hard-coded
- limit, throttling are hard-coded
- SecretCode, ClientID are coming in from flags/configuration file
- parameters:
  - state file for keeping refresh tokens
     - state file also includes all history downloaded
  - mindate is implicit


-->
 - For importing:
   --> Template allowed to entire JSON object.
   --> 
