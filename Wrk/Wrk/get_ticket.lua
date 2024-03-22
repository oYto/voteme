-- get_ticket.lua
wrk.method = "POST"
wrk.headers["Content-Type"] = "application/json"
wrk.body   = '{"query":"query { getCurrentTicket { ticketID validity } }"}'

