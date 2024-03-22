wrk.method = "POST"
wrk.headers["Content-Type"] = "application/json"
wrk.body   = '{"query":"query { getUserVotes(name: \\"Alice\\") }"}'

